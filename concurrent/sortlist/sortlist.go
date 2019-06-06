/*
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
Yet another Concurrent Skiplist implementation. Changes are performed using
atomic CAS operations. Inserts acquire a shared lock and Writes acquire an
exclusive lock. All locking is done internally.
*/
package sortlist

import (
	"fmt"
	"github.com/emirpasic/gods/utils"
	"unsafe"
	"sync/atomic"
	"math/rand"
	"sync"
)

type Node struct{
	Key   interface{}
	Value interface{}
	
	next  [64]unsafe.Pointer
	burnt bool
}
func (n *Node) String() string { return fmt.Sprint("{",n.Key," ",n.Value,"}") }
func (n *Node) get(i int) unsafe.Pointer {
	return atomic.LoadPointer(&n.next[i])
}
func (n *Node) set(i int,nw unsafe.Pointer) {
	atomic.StorePointer(&n.next[i],nw)
}
func (n *Node) cas(i int,old,nw unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&n.next[i],old,nw)
}
func (n *Node) Next() *Node {
	if n.burnt { return nil }
	return (*Node)(n.get(0))
}

type Sortlist struct{
	Cmp utils.Comparator
	Src rand.Source
	
	head Node
	rwm sync.RWMutex
}

func (s *Sortlist) icmp1(n *Node, sk interface{}) int {
	if n==nil { return 1 }
	return s.Cmp(n.Key,sk)
}

func (s *Sortlist) rand() int64 {
	if s.Src!=nil { return s.Src.Int63() }
	return rand.Int63()
}
func (s *Sortlist) randf() float64 {
	f := float64(s.rand()) / (1 << 63)
	return f
}

func (s *Sortlist) Previous(sk interface{}) *Node {
	var cur,past,elem *Node
	cur = &s.head
	
	for i := 63 ; i>=0 ; i-- {
		past = cur
		elem = (*Node)(past.get(i))
		for elem!=nil {
			if s.Cmp(elem.Key,sk) < 0 {
				past = elem
				elem = (*Node)(past.get(i))
			} else {
				break
			}
		}
		cur = past
	}
	if cur==&s.head { return nil }
	return cur
}
func (s *Sortlist) Floor(sk interface{}) *Node {
	var cur,past,elem *Node
	cur = &s.head
	
	for i := 63 ; i>=0 ; i-- {
		past = cur
		elem = (*Node)(past.get(i))
		for elem!=nil {
			if s.Cmp(elem.Key,sk) <= 0 {
				past = elem
				elem = (*Node)(past.get(i))
			} else {
				break
			}
		}
		cur = past
	}
	if cur==&s.head { return nil }
	return cur
}
func (s *Sortlist) Ceil(sk interface{}) *Node {
	n := s.Previous(sk)
	if n==nil { n = (*Node)(s.head.get(0)) }
	for n!=nil {
		if s.Cmp(n.Key,sk) >= 0 { return n }
		n = (*Node)(n.get(0))
	}
	return nil
}
func (s *Sortlist) Next(sk interface{}) *Node {
	n := s.Floor(sk)
	if n==nil { n = (*Node)(s.head.get(0)) }
	for n!=nil {
		if s.Cmp(n.Key,sk) > 0 { return n }
		n = (*Node)(n.get(0))
	}
	return nil
}
func (s *Sortlist) Lookup(sk interface{}) *Node {
	n := s.Floor(sk)
	if n==nil { return nil }
	if s.Cmp(n.Key,sk) > 0 { return nil }
	return n
}

func (s *Sortlist) Insert(k,v interface{}) {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	var cur,past,elem *Node
	cur = &s.head
	
	var backref [64]*Node
	
	for i := 63 ; i>=0 ; i-- {
		past = cur
		elem = (*Node)(past.get(i))
		for elem!=nil {
			if s.Cmp(elem.Key,k) <= 0 {
				past = elem
				elem = (*Node)(past.get(i))
			} else {
				break
			}
		}
		cur = past
		backref[i] = cur
	}
	
	prev := cur==&s.head
	if !prev { prev = s.Cmp(cur.Key,k) < 0 }
	
	if !prev {
		// XXX: We're treating an interface{}-field as atomic value.
		cur.Value = v
		return
	}
	
	cur = new(Node)
	cur.Key = k
	cur.Value = v
	
	f := s.randf()
	g := 0.5
	
	for i := 0 ; i<64 ; i++ {
		for {
			past = backref[i]
			elem = (*Node)(past.get(i))
			for elem !=nil {
				if s.Cmp(elem.Key,k) <= 0 {
					past = elem
					elem = (*Node)(past.get(i))
				} else {
					break
				}
			}
			cur.set(i,unsafe.Pointer(elem))
			if past.cas(i,unsafe.Pointer(elem),unsafe.Pointer(cur)) { break }
			cur.set(i,nil)
		}
		if f>g { break }
		g *= g
	}
}
func (s *Sortlist) Delete(k interface{}) *Node {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	var cur,past,elem *Node
	cur = &s.head
	
	var backref [64]*Node
	
	for i := 63 ; i>=0 ; i-- {
		past = cur
		elem = (*Node)(past.get(i))
		for elem!=nil {
			if s.Cmp(elem.Key,k) < 0 {
				past = elem
				elem = (*Node)(past.get(i))
			} else {
				break
			}
		}
		cur = past
		backref[i] = cur
	}
	cur = (*Node)(backref[0].get(0))
	
	if cur==&s.head { return nil }
	
	if s.Cmp(cur.Key,k) != 0 { return nil }
	
	for i := 1 ; i<64 ; i++ {
		if backref[i] != cur { backref[i] = nil }
	}
	
	for i := 0 ; i<64 ; i++ {
		if backref[i]==nil { continue }
		for {
			past = backref[i]
			elem = (*Node)(past.get(i))
			for elem !=nil && elem != cur {
				past = elem
				elem = (*Node)(past.get(i))
			}
			if elem == nil { break }
			repl := cur.get(i)
			if past.cas(i,unsafe.Pointer(cur),repl) { break }
		}
	}
	cur.burnt = true
	return cur
}

