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

package serializer

import "reflect"
import "github.com/byte-mug/golibs/preciseio"

type CodecElement interface{
	Read(r preciseio.PreciseReader,v reflect.Value) error
	Write(w *preciseio.PreciseWriter,v reflect.Value) error
}

type ceBlob struct{}
func (ce ceBlob) Read(r preciseio.PreciseReader,v reflect.Value) error {
	b,e := r.ReadBlob()
	if e!=nil { return e }
	v.SetBytes(b)
	return nil
}
func (ce ceBlob) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	return w.WriteBlob(v.Bytes())
}

type ceString struct{}
func (ce ceString) Read(r preciseio.PreciseReader,v reflect.Value) error {
	b,e := r.ReadBlob()
	if e!=nil { return e }
	v.SetString(string(b))
	return nil
}
func (ce ceString) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	return w.WriteBlob([]byte(v.String()))
}

type ceInt struct{}
func (ce ceInt) Read(r preciseio.PreciseReader,v reflect.Value) error {
	i,e := r.ReadVarint()
	if e!=nil { return e }
	v.SetInt(i)
	return nil
}
func (ce ceInt) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	return w.WriteVarint(v.Int())
}

type ceUint struct{}
func (ce ceUint) Read(r preciseio.PreciseReader,v reflect.Value) error {
	i,e := r.ReadUvarint()
	if e!=nil { return e }
	v.SetUint(i)
	return nil
}
func (ce ceUint) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	return w.WriteUvarint(v.Uint())
}

type ceSlice struct{
	child CodecElement
	t reflect.Type
}
func (ce ceSlice) Read(r preciseio.PreciseReader,v reflect.Value) error {
	n,e := r.ReadListLength()
	if e!=nil { return e }
	if n==0 {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	nv := reflect.MakeSlice(ce.t,n,n)
	for i:=0; i<n; i++ {
		e = ce.child.Read(r,nv.Index(i))
		if e!=nil { return e }
	}
	v.Set(nv)
	return nil
}
func (ce ceSlice) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	if v.IsNil() { return w.WriteListLength(0) }
	n := v.Len()
	e := w.WriteListLength(n)
	if e!=nil { return e }
	for i:=0; i<n; i++ {
		e = ce.child.Write(w,v.Index(i))
		if e!=nil { return e }
	}
	return nil
}


