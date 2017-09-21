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

// Extended IO Routines to construct Serializers/Deserializers
// to directly operate on bufio-Reader/bufio-Writer.
package preciseio

import "io"
import "encoding/binary"
import "errors"

var EListTooLong = errors.New("List too long")
var EBlobTooLong = errors.New("Blob too long")

const maxblob = (1<<24)-1

type Reader interface{
	io.Reader
	ReadByte() (byte, error)
}

type Writer interface{
	io.Writer
	WriteByte(c byte) error
}

//--------------------------------------------------------


type PreciseWriter struct{
	W Writer
	buf []byte
}
// This function initializes the internal fields.
// This must be called before any use.
func (pw *PreciseWriter) Initialize(){
	pw.buf = make([]byte,16)
}
func (pw PreciseWriter) WriteUvarint(i uint64) error {
	n := binary.PutUvarint(pw.buf,i)
	_,e := pw.W.Write(pw.buf[:n])
	return e
}
func (pw PreciseWriter) WriteVarint(i int64) error {
	n := binary.PutVarint(pw.buf,i)
	_,e := pw.W.Write(pw.buf[:n])
	return e
}
func (pw PreciseWriter) WriteBlob(b []byte) error {
	n := len(b)
	if n>maxblob { return EBlobTooLong }
	e := pw.WriteUvarint(uint64(n))
	if e!=nil { return e }
	_,e = pw.W.Write(b)
	return e
}
func (pw PreciseWriter) WriteListLength(i int) error {
	if i>maxblob { return EListTooLong }
	return pw.WriteUvarint(uint64(i))
}



type PreciseReader struct{
	R Reader
}
func (pr PreciseReader) ReadUvarint() (uint64,error) {
	return binary.ReadUvarint(pr.R)
}
func (pr PreciseReader) ReadVarint() (int64,error) {
	return binary.ReadVarint(pr.R)
}
func (pr PreciseReader) ReadBlob() ([]byte,error) {
	bn,e := pr.ReadUvarint()
	if e!=nil { return nil,e }
	n := int(bn&maxblob)
	if n==0 { return nil,nil }
	b := make([]byte,n)
	_,e = io.ReadFull(pr.R,b)
	if e!=nil { return nil,e }
	return b,nil
}
func (pr PreciseReader) ReadListLength() (int,error) {
	bn,e := pr.ReadUvarint()
	if e!=nil { return 0,e }
	n := int(bn&maxblob)
	return n,nil
}

