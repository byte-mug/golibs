/*
Copyright (c) 2018,2020 Simon Schmidt
Copyright (c) 2014 Armon Dadgar

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/



package radixrouter

import (
	"sort"
	"bytes"
	"sync"
	"regexp"
	"fmt"
)

func debug2(i ...interface{}) {
	i = append(append([]interface{}(nil),"radixrouter:"),i...)
	fmt.Println(i...)
}

// WalkFn is used when walking the tree. Takes a
// key and value, returning if iteration should
// be terminated.
type WalkFn func(s []byte, v interface{}) bool

type Parameter struct{
	Name  string
	Value []byte
}

type ParameterSetter func(key string, value interface{})

type paramList struct {
	next *paramList
	node *paramNode
	value []byte
	refc  int
}
var paramListPool = sync.Pool{New:func()interface{}{
	return new(paramList)
}}
func (p *paramList) unref() {
	var n *paramList
restart:
	if p==nil { return }
	p.refc--
	if p.refc>0 { return }
	n = p.next
	*p = paramList{}
	paramListPool.Put(p)
	p = n
	goto restart
}
func (p *paramList) ref() *paramList {
	if p==nil { return nil }
	p.refc++
	return p
}
func (p *paramList) add(node *paramNode,value []byte) *paramList {
	np := paramListPool.Get().(*paramList)
	*np = paramList{next:p.ref(),node:node,value:value,refc:1}
	return np
}
func (p* paramList) toParameter(buf []Parameter) (par []Parameter) {
	par = buf[:0]
	for i := p; i!=nil ; i = i.next {
		par = append(par,Parameter{i.node.name,i.value})
	}
	return
}
func (p* paramList) toParameterSetter(ps ParameterSetter) {
	for i := p; i!=nil ; i = i.next {
		ps(i.node.name,i.value)
	}
}


type paramNode struct {
	Tree
	name string
	cc   byte
}
func (p *paramNode) consume(search []byte) (prefix,rest []byte) {
	switch p.cc {
	case '*': return search,nil
	case ':':
		i := bytes.IndexByte(search,'/')
		if i<0 { return search,nil }
		return search[:i],search[i:]
	}
	return nil,search
}
func consumeSeg(suffix []byte) (prefix,rest []byte) {
	i := bytes.IndexByte(suffix,'/')
	if i<0 { return suffix,nil }
	return suffix[:i],suffix[i:]
}

// leafNode is used to represent a value
type leafNode struct {
	key []byte
	val interface{}
}

// edge is used to represent an edge node
type edge struct {
	label byte
	node  *node
}

type node struct {
	// leaf is used to store possible leaf
	leaf *leafNode
	
	// param is used to store possible parameter
	param *paramNode

	// prefix is the common prefix we ignore
	prefix []byte

	// Edges should be stored in-order for iteration.
	// We avoid a fully materialized slice to save memory,
	// since in most cases we expect to be sparse
	edges edges
}

func (n *node) isLeaf() bool {
	return n.leaf != nil
}
func (n *node) isParam() bool {
	return n.param != nil
}

func (n *node) addEdge(e edge) {
	n.edges = append(n.edges, e)
	n.edges.Sort()
}

func (n *node) updateEdge(label byte, node *node) {
	num := len(n.edges)
	idx := sort.Search(num, func(i int) bool {
		return n.edges[i].label >= label
	})
	if idx < num && n.edges[idx].label == label {
		n.edges[idx].node = node
		return
	}
	panic("replacing missing edge")
}

func (n *node) getEdge(label byte) *node {
	num := len(n.edges)
	idx := sort.Search(num, func(i int) bool {
		return n.edges[i].label >= label
	})
	if idx < num && n.edges[idx].label == label {
		return n.edges[idx].node
	}
	return nil
}

func (n *node) delEdge(label byte) {
	num := len(n.edges)
	idx := sort.Search(num, func(i int) bool {
		return n.edges[i].label >= label
	})
	if idx < num && n.edges[idx].label == label {
		copy(n.edges[idx:], n.edges[idx+1:])
		n.edges[len(n.edges)-1] = edge{}
		n.edges = n.edges[:len(n.edges)-1]
	}
}

type edges []edge

func (e edges) Len() int {
	return len(e)
}

func (e edges) Less(i, j int) bool {
	return e[i].label < e[j].label
}

func (e edges) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e edges) Sort() {
	sort.Sort(e)
}

// Tree implements a radix tree. This can be treated as a
// Dictionary abstract data type. The main advantage over
// a standard hash map is prefix-based lookups and
// ordered iteration,
type Tree struct {
	root *node
	size int
}

// New returns an empty Tree
func New() *Tree {
	return &Tree{root: &node{}}
}


// Len is used to return the number of elements in the tree
func (t *Tree) Len() int {
	return t.size
}

// longestPrefix finds the length of the shared prefix
// of two strings
func longestPrefix(k1, k2 []byte) int {
	max := len(k1)
	if l := len(k2); l < max {
		max = l
	}
	var i int
	for i = 0; i < max; i++ {
		if k1[i] != k2[i] {
			break
		}
	}
	return i
}

var paramStart = regexp.MustCompile(`\/[\:\*]`)

// InsertRoute insert a new route. If it overwrote the route, it returns
// (old-value,true) otherwise (nil,false).
func (t *Tree) InsertRoute(s []byte, v interface{}) (interface{}, bool) {
	for {
		loc := paramStart.FindIndex(s)
		if loc!=nil {
			cc := s[loc[0]+1]
			bname,rest := consumeSeg(s[loc[1]:])
			prefix := s[:loc[0]+1]
			t = t.InsertParameter(prefix,string(bname),cc)
			s = rest
			continue
		}
		return t.InsertRaw(s,v)
	}
}

// Maunually inserts a record into the radix tree.
func (t *Tree) InsertRaw(s []byte, v interface{}) (interface{}, bool) {
	var parent *node
	n := t.root
	search := s
	for {
		// Handle key exhaution
		if len(search) == 0 {
			if n.isLeaf() {
				old := n.leaf.val
				n.leaf.val = v
				return old, true
			}

			n.leaf = &leafNode{
				key: s,
				val: v,
			}
			t.size++
			return nil, false
		}

		// Look for the edge
		parent = n
		n = n.getEdge(search[0])

		// No edge, create one
		if n == nil {
			e := edge{
				label: search[0],
				node: &node{
					leaf: &leafNode{
						key: s,
						val: v,
					},
					prefix: search,
				},
			}
			parent.addEdge(e)
			t.size++
			return nil, false
		}

		// Determine longest prefix of the search key on match
		commonPrefix := longestPrefix(search, n.prefix)
		if commonPrefix == len(n.prefix) {
			search = search[commonPrefix:]
			continue
		}

		// Split the node
		t.size++
		child := &node{
			prefix: search[:commonPrefix],
		}
		parent.updateEdge(search[0], child)

		// Restore the existing node
		child.addEdge(edge{
			label: n.prefix[commonPrefix],
			node:  n,
		})
		n.prefix = n.prefix[commonPrefix:]

		// Create a new leaf node
		leaf := &leafNode{
			key: s,
			val: v,
		}

		// If the new key is a subset, add to to this node
		search = search[commonPrefix:]
		if len(search) == 0 {
			child.leaf = leaf
			return nil, false
		}

		// Create a new edge for the node
		child.addEdge(edge{
			label: search[0],
			node: &node{
				leaf:   leaf,
				prefix: search,
			},
		})
		return nil, false
	}
}

// Inserts a new Parameter and returns the subtree. If the subtree already existed
// AND name or cc differ, it panics.
func (t *Tree) InsertParameter(s []byte, name string, cc byte) (*Tree) {
	var parent *node
	n := t.root
	search := s
	for {
		// Handle key exhaution
		if len(search) == 0 {
			if n.isParam() {
				if n.param.name==name && n.param.cc==cc {
					return &(n.param.Tree)
				}
				panic(fmt.Sprintf("parameter conflict in trie: '%c%s' != '%c%s'",n.param.cc,n.param.name,cc,name))
				return nil
			}

			n.param = &paramNode{
				name: name,
				cc: cc,
			}
			n.param.root = new(node)
			t.size++
			return &(n.param.Tree)
		}

		// Look for the edge
		parent = n
		n = n.getEdge(search[0])

		param := &paramNode{
			name: name,
			cc: cc,
		}
		param.root = new(node)
		
		// No edge, create one
		if n == nil {
			
			e := edge{
				label: search[0],
				node: &node{
					param: param,
					prefix: search,
				},
			}
			parent.addEdge(e)
			t.size++
			return &(param.Tree)
		}

		// Determine longest prefix of the search key on match
		commonPrefix := longestPrefix(search, n.prefix)
		if commonPrefix == len(n.prefix) {
			search = search[commonPrefix:]
			continue
		}

		// Split the node
		t.size++
		child := &node{
			prefix: search[:commonPrefix],
		}
		parent.updateEdge(search[0], child)

		// Restore the existing node
		child.addEdge(edge{
			label: n.prefix[commonPrefix],
			node:  n,
		})
		n.prefix = n.prefix[commonPrefix:]

		// If the new key is a subset, add to to this node
		search = search[commonPrefix:]
		if len(search) == 0 {
			child.param = param
			return &(param.Tree)
		}

		// Create a new edge for the node
		child.addEdge(edge{
			label: search[0],
			node: &node{
				param: param,
				prefix: search,
			},
		})
		return &(param.Tree)
	}
}

// Delete is used to delete a key, returning the previous
// value and if it was deleted
func (t *Tree) Delete(s []byte) (interface{}, bool) {
	var parent *node
	var label byte
	n := t.root
	search := s
	for {
		// Check for key exhaution
		if len(search) == 0 {
			if !n.isLeaf() {
				break
			}
			goto DELETE
		}

		// Look for an edge
		parent = n
		label = search[0]
		n = n.getEdge(label)
		if n == nil {
			break
		}

		// Consume the search prefix
		if bytes.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
	return nil, false

DELETE:
	// Delete the leaf
	leaf := n.leaf
	n.leaf = nil
	t.size--

	// Check if we should delete this node from the parent
	if parent != nil && len(n.edges) == 0 {
		parent.delEdge(label)
	}

	// Check if we should merge this node
	if n != t.root && len(n.edges) == 1 {
		n.mergeChild()
	}

	// Check if we should merge the parent's other child
	if parent != nil && parent != t.root && len(parent.edges) == 1 && !parent.isLeaf() {
		parent.mergeChild()
	}

	return leaf.val, true
}

// DeletePrefix is used to delete the subtree under a prefix
// Returns how many nodes were deleted
// Use this to delete large subtrees efficiently
func (t *Tree) DeletePrefix(s []byte) int {
	return t.deletePrefix(nil, t.root, s)
}

// delete does a recursive deletion
func (t *Tree) deletePrefix(parent, n *node, prefix []byte) int {
	// Check for key exhaustion
	if len(prefix) == 0 {
		// Remove the leaf node
		subTreeSize := 0
		//recursively walk from all edges of the node to be deleted
		recursiveWalk(n, func(s []byte, v interface{}) bool {
			subTreeSize++
			return false
		})
		if n.isLeaf() {
			n.leaf = nil
		}
		n.edges = nil // deletes the entire subtree

		// Check if we should merge the parent's other child
		if parent != nil && parent != t.root && len(parent.edges) == 1 && !parent.isLeaf() {
			parent.mergeChild()
		}
		t.size -= subTreeSize
		return subTreeSize
	}

	// Look for an edge
	label := prefix[0]
	child := n.getEdge(label)
	if child == nil || (!bytes.HasPrefix(child.prefix, prefix) && !bytes.HasPrefix(prefix, child.prefix)) {
		return 0
	}

	// Consume the search prefix
	if len(child.prefix) > len(prefix) {
		prefix = prefix[len(prefix):]
	} else {
		prefix = prefix[len(child.prefix):]
	}
	return t.deletePrefix(n, child, prefix)
}

func (n *node) mergeChild() {
	e := n.edges[0]
	child := e.node
	// TODO: make this better!
	n.prefix = []byte(string(n.prefix) + string(child.prefix))
	n.leaf = child.leaf
	n.edges = child.edges
}

// GetParlist is used to perform a lookup in the route.
func (t *Tree) GetParlist(s []byte, par []Parameter) (rv interface{}, rpar []Parameter, rok bool) {
	var lpar *paramList
	rv,lpar,rok = t.get(s,lpar)
	rpar = lpar.toParameter(par)
	lpar.unref()
	return
}

// Get is used to perform a lookup in the route.
func (t *Tree) Get(s []byte, ps ParameterSetter) (rv interface{}, rok bool) {
	var lpar *paramList
	rv,lpar,rok = t.get(s,lpar)
	if ps!=nil { lpar.toParameterSetter(ps) }
	lpar.unref()
	return
}


func (t *Tree) get(s []byte, par *paramList) (rv interface{}, rpar *paramList, rok bool) {
	n := t.root
	search := s
	for {
		// Check for key exhaution
		if len(search) == 0 {
			if n.isLeaf() {
				rpar.unref()
				return n.leaf.val, par, true
			}
			break
		}
		
		if n.isParam() {
			p,r := n.param.consume(search)
			xv,xpar,xok := n.param.get(r,par)
			if xok {
				rpar.unref()
				rv,rpar,rok = xv,xpar.add(n.param,p),true
			}
		}

		// Look for an edge
		n = n.getEdge(search[0])
		if n == nil {
			break
		}

		// Consume the search prefix
		if bytes.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
	return
}


// LongestPrefix is like Get, but instead of an
// exact match, it will return the longest prefix match.
func (t *Tree) LongestPrefix(s []byte) ([]byte, interface{}, bool) {
	var last *leafNode
	n := t.root
	search := s
	for {
		// Look for a leaf node
		if n.isLeaf() {
			last = n.leaf
		}

		// Check for key exhaution
		if len(search) == 0 {
			break
		}

		// Look for an edge
		n = n.getEdge(search[0])
		if n == nil {
			break
		}

		// Consume the search prefix
		if bytes.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
	if last != nil {
		return last.key, last.val, true
	}
	return []byte{}, nil, false
}

// Minimum is used to return the minimum value in the tree
func (t *Tree) Minimum() ([]byte, interface{}, bool) {
	n := t.root
	for {
		if n.isLeaf() {
			return n.leaf.key, n.leaf.val, true
		}
		if len(n.edges) > 0 {
			n = n.edges[0].node
		} else {
			break
		}
	}
	return []byte{}, nil, false
}

// Maximum is used to return the maximum value in the tree
func (t *Tree) Maximum() ([]byte, interface{}, bool) {
	n := t.root
	for {
		if num := len(n.edges); num > 0 {
			n = n.edges[num-1].node
			continue
		}
		if n.isLeaf() {
			return n.leaf.key, n.leaf.val, true
		}
		break
	}
	return []byte{}, nil, false
}

// Walk is used to walk the tree
func (t *Tree) Walk(fn WalkFn) {
	recursiveWalk(t.root, fn)
}

// WalkPrefix is used to walk the tree under a prefix
func (t *Tree) WalkPrefix(prefix []byte, fn WalkFn) {
	n := t.root
	search := prefix
	for {
		// Check for key exhaution
		if len(search) == 0 {
			recursiveWalk(n, fn)
			return
		}

		// Look for an edge
		n = n.getEdge(search[0])
		if n == nil {
			break
		}

		// Consume the search prefix
		if bytes.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]

		} else if bytes.HasPrefix(n.prefix, search) {
			// Child may be under our search prefix
			recursiveWalk(n, fn)
			return
		} else {
			break
		}
	}

}

// WalkPath is used to walk the tree, but only visiting nodes
// from the root down to a given leaf. Where WalkPrefix walks
// all the entries *under* the given prefix, this walks the
// entries *above* the given prefix.
func (t *Tree) WalkPath(path []byte, fn WalkFn) {
	n := t.root
	search := path
	for {
		// Visit the leaf values if any
		if n.leaf != nil && fn(n.leaf.key, n.leaf.val) {
			return
		}

		// Check for key exhaution
		if len(search) == 0 {
			return
		}

		// Look for an edge
		n = n.getEdge(search[0])
		if n == nil {
			return
		}

		// Consume the search prefix
		if bytes.HasPrefix(search, n.prefix) {
			search = search[len(n.prefix):]
		} else {
			break
		}
	}
}

// recursiveWalk is used to do a pre-order walk of a node
// recursively. Returns true if the walk should be aborted
func recursiveWalk(n *node, fn WalkFn) bool {
	// Visit the leaf values if any
	if n.leaf != nil && fn(n.leaf.key, n.leaf.val) {
		return true
	}

	// Recurse on the children
	for _, e := range n.edges {
		if recursiveWalk(e.node, fn) {
			return true
		}
	}
	return false
}
