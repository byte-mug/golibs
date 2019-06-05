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

type NodeID []byte

/*
Assigns the Value of the NodeID 'other' to 'n', adjusting the buffer size, if necessary.

Internally, it does:

	*n = append((*n)[:0],other...)
*/
func (n *NodeID) Set(other NodeID) { *n = append((*n)[:0],other...) }

/*
Generates a Copy.
*/
func (n NodeID) Clone() NodeID { return append(make(NodeID,0,len(n)),n...) }

/*
Returns:

	len(n)*8
*/
func (n NodeID) Bits() int { return len(n)*8 }

/*
Calculates:

	n = (n + (1<<k)) mod 1<<n.Bits()
*/
func (n NodeID) FingerBase(k uint) {
	if k>=uint(n.Bits()) { return }
	inv := len(n)-1
	
	i := int(k>>3)
	r := uint16(n[inv-i]) + uint16(1<<(k&7))
	n[inv-i] = uint8(r&0xff)
	if r<=0xff { return }
	for i++ ; i < len(n) ; i++ {
		n[inv-i]++
		if n[inv-i]!=0 { break }
	}
}

/*
Calculates:

	n = (n + 1) mod 1<<n.Bits()
*/
func (n NodeID) Increment() {
	inv := len(n)-1
	for i := 0 ; i < len(n) ; i++ {
		n[inv-i]++
		if n[inv-i]!=0 { break }
	}
}

/*
Calculates:

	n = (n - 1) mod 1<<n.Bits()
*/
func (n NodeID) Decrement() {
	inv := len(n)-1
	for i := 0 ; i < len(n) ; i++ {
		n[inv-i]--
		if n[inv-i]!=0xff { break }
	}
}
