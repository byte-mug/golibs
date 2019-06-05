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
*/
package skiplist

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
	"github.com/emirpasic/gods/utils"
)

const (
	f_first uint = 1<<iota
)

type Skipnode struct {
	Key     interface{}
	Val     interface{}
	Forward []*Skipnode
	Level   int
	
	flags   uint
}
func (s *Skipnode) compare(c utils.Comparator,key interface{}) int {
	if (s.flags&f_first)==0 { return c(s.Key,key) }
	return -1
}

func NewNode(searchKey interface{}, value interface{}, createLevel int, maxLevel int) *Skipnode {
	//Every forward prepare a maxLevel empty point first.
	forwardEmpty := make([]*Skipnode, maxLevel)
	for i := 0; i <= maxLevel-1; i++ {
		forwardEmpty[i] = nil
	}
	return &Skipnode{Key: searchKey, Val: value, Forward: forwardEmpty, Level: createLevel}
}
func NewHeader(createLevel int, maxLevel int) *Skipnode {
	//Every forward prepare a maxLevel empty point first.
	forwardEmpty := make([]*Skipnode, maxLevel)
	for i := 0; i <= maxLevel-1; i++ {
		forwardEmpty[i] = nil
	}
	return &Skipnode{Forward: forwardEmpty, Level: createLevel, flags:f_first}
}

type Skiplist struct {
	Header *Skipnode
	
	// List configuration
	MaxLevel    int
	Propobility float32
	
	// List Comparison
	Comparator  utils.Comparator
	
	// List status
	Level int //current level of whole skiplist
}

const (
	DefaultMaxLevel    int     = 15   //Maximal level allow to create in this skip list
	DefaultPropobility float32 = 0.25 //Default propobility
)

//NewSkipList : Init structure for Skit List.
func NewSkipList(comparator utils.Comparator) *Skiplist {
	newList := &Skiplist{Header: NewHeader(1, DefaultMaxLevel), Comparator: comparator, Level: 1}
	newList.MaxLevel = DefaultMaxLevel       //default
	newList.Propobility = DefaultPropobility //default
	return newList
}

func randomP() float32 {
	rand.Seed(int64(time.Now().Nanosecond()))
	return rand.Float32()
}

//Change SkipList default maxlevel is 4.
func (b *Skiplist) SetMaxLevel(maxLevel int) {
	b.MaxLevel = maxLevel
}

func (b *Skiplist) RandomLevel() int {
	level := 1
	for randomP() < b.Propobility && level < b.MaxLevel {
		level++
	}
	return level
}

//Search: Search a element by search key and return the interface{}
func (b *Skiplist) Search(searchKey interface{}) (interface{}, error) {
	currentNode := b.Header

	//Start traversal forward first.
	for i := b.Level - 1; i >= 0; i-- {
		for currentNode.Forward[i] != nil && currentNode.Forward[i].compare(b.Comparator,searchKey) < 0 {
			currentNode = currentNode.Forward[i]
		}
	}

	//Step to final search node.
	currentNode = currentNode.Forward[0]

	if currentNode != nil && currentNode.compare(b.Comparator,searchKey) == 0 {
		return currentNode.Val, nil
	}
	return nil, errors.New("Not found.")
}

//Insert: Insert a search key and its value which could be interface.
func (b *Skiplist) Insert(searchKey interface{}, value interface{}) {
	updateList := make([]*Skipnode, b.MaxLevel)
	currentNode := b.Header

	//Quick search in forward list
	for i := b.Header.Level - 1; i >= 0; i-- {
		for currentNode.Forward[i] != nil && currentNode.Forward[i].compare(b.Comparator,searchKey) < 0 {
			currentNode = currentNode.Forward[i]
		}
		updateList[i] = currentNode
	}

	//Step to next node. (which is the target insert location)
	currentNode = currentNode.Forward[0]

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
			newNode.Forward[i] = updateList[i].Forward[i]
			updateList[i].Forward[i] = newNode
		}
	}
}

//Delete: Delete element by search key
func (b *Skiplist) Delete(searchKey interface{}) error {
	updateList := make([]*Skipnode, b.MaxLevel)
	currentNode := b.Header

	//Quick search in forward list
	for i := b.Header.Level - 1; i >= 0; i-- {
		for currentNode.Forward[i] != nil && currentNode.Forward[i].compare(b.Comparator,searchKey) < 0 {
			currentNode = currentNode.Forward[i]
		}
		updateList[i] = currentNode
	}

	//Step to next node. (which is the target delete location)
	currentNode = currentNode.Forward[0]

	if currentNode.compare(b.Comparator,searchKey) == 0 {
		for i := 0; i <= currentNode.Level-1; i++ {
			if updateList[i].Forward[i] != nil && updateList[i].Forward[i].compare(b.Comparator,currentNode.Key) != 0 {
				break
			}
			updateList[i].Forward[i] = currentNode.Forward[i]
		}

		for currentNode.Level > 1 && b.Header.Forward[currentNode.Level] == nil {
			currentNode.Level--
		}

		//free(currentNode)  //no need for Golang because GC
		currentNode = nil
		return nil
	}
	return errors.New("Not found")
}

//DisplayAll: Display current SkipList content in console, will also print out the linked pointer.
func (b *Skiplist) DisplayAll() {
	fmt.Printf("\nhead->")
	currentNode := b.Header

	//Draw forward[0] base
	for {
		if (currentNode.flags&f_first)==0 {
			fmt.Printf("[key:%v][val:%v]->", currentNode.Key, currentNode.Val)
		} else {
			fmt.Printf("[HEAD]->")
		}
		if currentNode.Forward[0] == nil {
			break
		}
		currentNode = currentNode.Forward[0]
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

		if currentNode.Forward[0] == nil {
			break
		}

		for j := currentNode.Level - 1; j >= 0; j-- {
			fmt.Printf(" fw[%d]:", j)
			if currentNode.Forward[j] != nil {
				fmt.Printf("%v", currentNode.Forward[j].Key)
			} else {
				fmt.Printf("nil")
			}
		}
		fmt.Printf("\n")
		currentNode = currentNode.Forward[0]
	}
	fmt.Printf("\n")
}
