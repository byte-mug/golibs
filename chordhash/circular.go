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


package chordhash

import (
	avl "github.com/emirpasic/gods/trees/avltree"
	"fmt"
)

type Circle struct{
	Tree *avl.Tree
}
func (c* Circle) Init() {
	c.Tree = avl.NewWithStringComparator()
}
func (c* Circle) Next(key interface{}) (node *avl.Node) {
	node,_ = c.Tree.Floor(key)
	// Lemma: if node!=nil then node.Key <= key
	
	if node!=nil { node = node.Next() }
	// Lemma: if node!=nil then key < node.Key
	
	if node==nil { node = c.Tree.Left() }
	return
}
func (c* Circle) Prev(key interface{}) (node *avl.Node) {
	node,_ = c.Tree.Ceiling(key)
	// Lemma: if node!=nil then key <= node.Key
	
	if node!=nil { node = node.Prev() }
	// Lemma: if node!=nil then node.Key < key
	
	if node==nil { node = c.Tree.Right() }
	return
}
func (c *Circle) NextOrEqual(key interface{}) (node *avl.Node) {
	node,_ = c.Tree.Ceiling(key)
	
	if node==nil { node = c.Tree.Left() }
	
	return
}
func (c *Circle) PrevOrEqual(key interface{}) (node *avl.Node) {
	node,_ = c.Tree.Floor(key)
	
	if node==nil { node = c.Tree.Right() }
	
	return
}

func (c *Circle) Step(node *avl.Node) (next *avl.Node) {
	next = node.Next()
	if next==nil { next = c.Tree.Left() }
	return
}

func (c *Circle) StepReverse(node *avl.Node) (prev *avl.Node) {
	prev = node.Prev()
	if prev==nil { prev = c.Tree.Right() }
	return
}

func NodeNonNil(n *avl.Node) *avl.Node {
	if n==nil { n = new(avl.Node) }
	return n
}
func Strnode(n *avl.Node) string {
	if n==nil { return "<nil>" }
	return fmt.Sprintf("%x -> %v",n.Key,n.Value)
}

