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
)

/*
A Node manager, offering a holistic member-list and -ring implementation.

This implementation requires the whole list of cluster-members, whereas Chord
only requires a Finger-Table.
*/
type NodeManager struct{
	Finger,Nodes Circle
	Self string
}
func (n *NodeManager) Init() {
	n.Finger.Init()
	n.Nodes.Init()
}
func (n *NodeManager) SetSelf(id NodeID) {
	var nid NodeID
	bits := id.Bits()
	
	n.Finger.Tree.Clear()
	
	for i := 0 ; i < bits ; i++ {
		nid.Set(id)
		nid.FingerBase(uint(i))
		n.Finger.Tree.Put(string(nid),i)
	}
}
func (n *NodeManager) Insert(id NodeID, attachment interface{}) {
	n.Nodes.Tree.Put(string(id),attachment)
}
func (n *NodeManager) Replace(id NodeID, new_attachment interface{}) {
	n.Nodes.Tree.Put(string(id),new_attachment)
}
func (n *NodeManager) Remove(id NodeID) {
	n.Nodes.Tree.Remove(string(id))
}
func (n *NodeManager) LookupPrecise(id NodeID) *avl.Node {
	return n.Nodes.Next(string(id))
}
func (n *NodeManager) LookupFinger(id NodeID) *avl.Node {
	fino := n.Finger.PrevOrEqual(string(id))
	if fino==nil { return nil }
	return n.Nodes.Next(fino.Key)
}
func (n *NodeManager) Successor() *avl.Node {
	return n.Nodes.Next(n.Self)
}
func (n *NodeManager) Predecessor() *avl.Node {
	return n.Nodes.Prev(n.Self)
}


/**/
