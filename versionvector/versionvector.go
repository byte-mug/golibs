/*
MIT License

Copyright (c) 2017 Simon Schmidt

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
An implementation of the Version vector technique for distributed systems.


Basics

A Version vector is a set of key value pairs, where the key IDENTIFIES the Node
modifying the data item, and the value is the version. If a key value-pair with
a given key does not exist, 0 is the default value.


Requirements

Two Nodes MUST NOT have the same key and the key SHOULD NOT change over the
lifetime of the Node.


Caveats

The algorithm is capable to detect Conflicts but it is not capable to resolve
conflicts that occured.
*/
package versionvector

type Comparison uint
const (
	comp_same Comparison = 0
	comp_lower Comparison = 1
	comp_higher Comparison = 2
	comp_conflict = comp_lower|comp_higher
)
func compareVersion(a, b uint64) Comparison {
	if a<b { return comp_lower }
	if a>b { return comp_higher }
	return comp_same
}
func GetComparison(lower, higher bool) Comparison {
	s := comp_same
	if lower { s|=comp_lower }
	if higher { s|=comp_higher }
	return s
}
func (c Comparison) Lower() bool { return c==comp_lower }
func (c Comparison) Higher() bool { return c==comp_higher }
func (c Comparison) Conflict() bool { return c==comp_conflict }
func (c Comparison) AsInt() int {
	if c==comp_lower { return -1 }
	if c==comp_higher { return 1 }
	return 0
}
func (c Comparison) String() string {
	switch c {
	case comp_same: return "Same"
	case comp_lower: return "Lower"
	case comp_higher: return "Higher"
	case comp_conflict: return "Conflict"
	}
	return "???"
}

/* A Version vector with strings as node-IDs. */
type Vector map[string]uint64
func (v Vector) Increment(node string) { v[node] = (v[node])+1 }
func (v Vector) Compare(other Vector) Comparison {
	s := comp_same
	for k := range v { s|= compareVersion(v[k],other[k]) }
	for k := range other { s|= compareVersion(v[k],other[k]) }
	return s
}
func (v Vector) Clone() Vector {
	nv := make(Vector,len(v))
	for k,vv := range v { nv[k]=vv }
	return nv
}

/* A Version vector with integers as node-IDs. */
type IntVector []uint64

/* Increments the node's version and returns the new vector. */
func (v IntVector) Increment(node uint64) IntVector {
	L := len(v)&^1
	nv := make(IntVector,0,L+2)
	
	i := 0
	for ; i<L ; i+=2 {
		if v[i]>=node { break }
		nv = append(nv,v[i],v[i+1])
	}
	if i<L && v[i]==node {
		nv = append(nv,node,v[i+1]+1)
		i+=2
	}else{
		nv = append(nv,node,1)
	}
	for ; i<L ; i+=2 {
		nv = append(nv,v[i],v[i+1])
	}
	return nv
}
const eolist = ^uint64(0)
func iseolist(k,r uint64) bool { return k==eolist&&r==0 }
func (v *IntVector) next() (k,r uint64) {
	if len(*v)<2 { return eolist,0 }
	k = (*v)[0]
	r = (*v)[1]
	*v = (*v)[2:]
	return
}
func (v IntVector) Compare(other IntVector) Comparison {
	s := comp_same
	Ak,Ar := v.next()
	Bk,Br := other.next()
	for !(iseolist(Ak,Ar)&&iseolist(Bk,Br)){
		if Ak<Bk { // Bk is begind Ak, so Br==0
			s |= compareVersion(Ar,0)
			Ak,Ar = v.next()
		} else if Ak>Bk { // Ak is behind Bk, so Ar==0
			s |= compareVersion(0,Br)
			Bk,Br = other.next()
		} else {
			s |= compareVersion(Ar,Br)
			Ak,Ar = v.next()
			Bk,Br = other.next()
		}
	}
	return s
}


