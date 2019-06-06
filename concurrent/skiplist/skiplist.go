/*
Copyright (c) 2015 Evan Lin
Copyright (c) 2019 Simon Schmidt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/


/*
Skip List implement in Go.

A Skip List is a data structure that allows fast search within an ordered
sequence of elements. Fast search is made possible by maintaining a linked
hierarchy of subsequences, each skipping over fewer elements.

This Implementation is Partially Concurrent: It supports a single writer
and multiple concurrent readers. Readers can access the Skiplist without
locking.
*/
package skiplist

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
	"github.com/emirpasic/gods/utils"
	"unsafe"
	"sync/atomic"
)

const (
	f_first uint = 1<<iota
)

var (
	ErrNotFound = errors.New("Not found.")
)

type Skipnode struct {
	Key     interface{}
	Val     interface{}
	aForward []unsafe.Pointer
	Level   int
	
	flags   uint
}
func (s *Skipnode) compare(c utils.Comparator,key interface{}) int {
	if (s.flags&f_first)==0 { return c(s.Key,key) }
	return -1
}
func (s *Skipnode) rawForward(i int) unsafe.Pointer {
	return atomic.LoadPointer(&(s.aForward[i]))
}
func (s *Skipnode) forward(i int) *Skipnode {
	return (*Skipnode)(atomic.LoadPointer(&(s.aForward[i])))
}
func (s *Skipnode) setForward(i int, sn *Skipnode) {
	atomic.StorePointer(&(s.aForward[i]),unsafe.Pointer(s))
}
func (s *Skipnode) forwardFrom(i int,o *Skipnode, j int) {
	atomic.StorePointer(&(s.aForward[i]),o.rawForward(j))
}



func NewNode(searchKey interface{}, value interface{}, createLevel int, maxLevel int) *Skipnode {
	//Every forward prepare a maxLevel empty point first.
	forwardEmpty := make([]unsafe.Pointer, maxLevel)
	for i := 0; i <= maxLevel-1; i++ {
		forwardEmpty[i] = nil
	}
	return &Skipnode{Key: searchKey, Val: value, aForward: forwardEmpty, Level: createLevel}
}
func NewHeader(createLevel int, maxLevel int) *Skipnode {
	//Every forward prepare a maxLevel empty point first.
	forwardEmpty := make([]unsafe.Pointer, maxLevel)
	for i := 0; i <= maxLevel-1; i++ {
		forwardEmpty[i] = nil
	}
	return &Skipnode{aForward: forwardEmpty, Level: createLevel, flags:f_first}
}

type Skiplist struct {
	Header *Skipnode
	
	// List configuration
	MaxLevel    int
	Propability float32
	
	// List Comparison
	Comparator  utils.Comparator
	
	// List status
	Level int //current level of whole skiplist
	
	Random *rand.Rand
}

const (
	DefaultMaxLevel    int     = 15   //Maximal level allow to create in this skip list
	DefaultPropability float32 = 0.25 //Default Propability
)

//NewSkipList : Init structure for Skit List.
func NewSkipList(comparator utils.Comparator) *Skiplist {
	newList := &Skiplist{Header: NewHeader(1, DefaultMaxLevel), Comparator: comparator, Level: 1}
	newList.MaxLevel = DefaultMaxLevel       //default
	newList.Propability = DefaultPropability //default
	newList.Random = rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	return newList
}

func (b *Skiplist) randomP() float32 {
	if b.Random == nil {
		b.Random = rand.New(rand.NewSource(rand.Int63()))
	}
	return b.Random.Float32()
}

//Change SkipList default maxlevel is 4.
func (b *Skiplist) SetMaxLevel(maxLevel int) {
	b.MaxLevel = maxLevel
}

func (b *Skiplist) RandomLevel() int {
	level := 1
	for b.randomP() < b.Propability && level < b.MaxLevel {
		level++
	}
	return level
}

//Search: Search a element by search key and return the interface{}
func (b *Skiplist) Search(searchKey interface{}) (interface{}, error) {
	currentNode := b.Header

	//Start traversal forward first.
	for i := b.Level - 1; i >= 0; i-- {
		for currentNode.rawForward(i) != nil && currentNode.forward(i).compare(b.Comparator,searchKey) < 0 {
			currentNode = currentNode.forward(i)
		}
	}

	//Step to final search node.
	currentNode = currentNode.forward(0)

	if currentNode != nil && currentNode.compare(b.Comparator,searchKey) == 0 {
		return currentNode.Val, nil
	}
	return nil, ErrNotFound
}

//Insert: Insert a search key and its value which could be interface.
func (b *Skiplist) Insert(searchKey interface{}, value interface{}) {
	updateList := make([]*Skipnode, b.MaxLevel)
	currentNode := b.Header

	//Quick search in forward list
	for i := b.Header.Level - 1; i >= 0; i-- {
		for currentNode.rawForward(i) != nil && currentNode.forward(i).compare(b.Comparator,searchKey) < 0 {
			currentNode = currentNode.forward(i)
		}
		updateList[i] = currentNode
	}

	//Step to next node. (which is the target insert location)
	currentNode = currentNode.forward(0)

	if currentNode != nil && currentNode.compare(b.Comparator,searchKey) == 0 {
		currentNode.Val = value
	} else {
		newLevel := b.RandomLevel()
		if newLevel > b.Level {
			for i := b.Level + 1; i <= newLevel; i++ {
				updateList[i-1] = b.Header
			}
			b.Level = newLevel //This is not mention in cookbook pseudo code
			b.Header.Level = newLevel
		}

		newNode := NewNode(searchKey, value, newLevel, b.MaxLevel) //New node
		for i := 0; i <= newLevel-1; i++ {                         //zero base
			newNode.forwardFrom(i,updateList[i],i)
		}
		
		// XXXMFG: I need a write barrier (eg. Write-Cache-Flush) here badly!!!
		
		for i := 0; i <= newLevel-1; i++ {                         //zero base
			updateList[i].setForward(i,newNode)
		}
		
		// We assume, that
		//	newNode.Forward[i] = updateList[i].Forward[i]
		// is observable before
		//	updateList[i].Forward[i] = newNode
		//
		// If not, !BANG!
	}
}

//Delete: Delete element by search key
func (b *Skiplist) Delete(searchKey interface{}) error {
	updateList := make([]*Skipnode, b.MaxLevel)
	currentNode := b.Header

	//Quick search in forward list
	for i := b.Header.Level - 1; i >= 0; i-- {
		for currentNode.rawForward(i) != nil && currentNode.forward(i).compare(b.Comparator,searchKey) < 0 {
			currentNode = currentNode.forward(i)
		}
		updateList[i] = currentNode
	}

	//Step to next node. (which is the target delete location)
	currentNode = currentNode.forward(0)

	if currentNode.compare(b.Comparator,searchKey) == 0 {
		for i := 0; i <= currentNode.Level-1; i++ {
			if updateList[i].rawForward(i) != nil && updateList[i].forward(i).compare(b.Comparator,currentNode.Key) != 0 {
				break
			}
			updateList[i].forwardFrom(i,currentNode,i)
		}

		for currentNode.Level > 1 && b.Header.rawForward(currentNode.Level) == nil {
			currentNode.Level--
		}

		//free(currentNode)  //no need for Golang because GC
		currentNode = nil
		return nil
	}
	return ErrNotFound
}

//DisplayAll: Display current SkipList content in console, will also print out the linked pointer.
func (b *Skiplist) DisplayAll() {
	fmt.Printf("\nhead->")
	currentNode := b.Header

	//Draw forward[0] base
	for {
		if (currentNode.flags&f_first)==0 {
			fmt.Printf("[key:%v][val:%v]->\n\t", currentNode.Key, currentNode.Val)
		} else {
			fmt.Printf("[HEAD]->\n\t")
		}
		if currentNode.rawForward(0) == nil {
			break
		}
		currentNode = currentNode.forward(0)
	}
	fmt.Printf("nil\n")

	fmt.Println("---------------------------------------------------------")
	currentNode = b.Header
	//Draw all data node.
	for {
		if (currentNode.flags&f_first)==0 {
			fmt.Printf("[node:%v], val:%v, level:%d ", currentNode.Key, currentNode.Val, currentNode.Level)
		} else {
			fmt.Printf("[HEAD], level:%d ", currentNode.Level)
		}

		if currentNode.rawForward(0) == nil {
			break
		}

		for j := currentNode.Level - 1; j >= 0; j-- {
			fmt.Printf(" fw[%d]:", j)
			if currentNode.rawForward(j) != nil {
				fmt.Printf("%v", currentNode.forward(j).Key)
			} else {
				fmt.Printf("nil")
			}
		}
		fmt.Printf("\n")
		currentNode = currentNode.forward(0)
	}
	fmt.Printf("\n")
}
