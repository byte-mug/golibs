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

type ceByte struct{}
func (ce ceByte) Read(r preciseio.PreciseReader,v reflect.Value) error {
	b,e := r.R.ReadByte()
	if e!=nil { return e }
	v.SetUint(uint64(b))
	return nil
}
func (ce ceByte) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	return w.W.WriteByte(byte(v.Uint()))
}

type ceSbyte struct{}
func (ce ceSbyte) Read(r preciseio.PreciseReader,v reflect.Value) error {
	b,e := r.R.ReadByte()
	if e!=nil { return e }
	i := int8(b)
	v.SetInt(int64(i))
	return nil
}
func (ce ceSbyte) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	i := int8(v.Int())
	return w.W.WriteByte(byte(i))
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
	v = CastV(ce.t,v)
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

type ceMap struct{
	k CodecElement
	v CodecElement
	t reflect.Type
}
func (ce ceMap) Read(r preciseio.PreciseReader,v reflect.Value) error {
	b,e := r.R.ReadByte()
	if e!=nil { return e }
	if b==0 {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	n,e := r.ReadListLength()
	if e!=nil { return e }
	nv := reflect.MakeMapWithSize(ce.t,n)
	ckv := reflect.New(ce.t.Key()).Elem()
	cvv := reflect.New(ce.t.Elem()).Elem()
	for i:=0 ; i<n; i++ {
		e = ce.k.Read(r,ckv)
		if e!=nil { return e }
		e = ce.v.Read(r,cvv)
		if e!=nil { return e }
		nv.SetMapIndex(ckv,cvv)
	}
	v.Set(nv)
	return nil
}
func (ce ceMap) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	v = CastV(ce.t,v)
	if v.IsNil() { return w.W.WriteByte(0) }
	e := w.W.WriteByte(0xff)
	if e!=nil { return e }
	keys := v.MapKeys()
	n := len(keys)
	e = w.WriteListLength(n)
	if e!=nil { return e }
	for _,key := range keys {
		e = ce.k.Write(w,key)
		if e!=nil { return e }
		e = ce.v.Write(w,v.MapIndex(key))
		if e!=nil { return e }
	}
	return nil
}

type ceArray struct{
	child CodecElement
	t reflect.Type
}
func (ce ceArray) Read(r preciseio.PreciseReader,v reflect.Value) error {
	n := ce.t.Len()
	nv := v
	wrongtype := v.Type()!=ce.t // Type-mismatch
	settable := v.CanSet() // Non-Settable
	if wrongtype || !settable { nv = reflect.New(ce.t).Elem() } // Allocate in heap.
	for i:=0; i<n; i++ {
		e := ce.child.Read(r,nv.Index(i))
		if e!=nil { return e }
	}
	if wrongtype || !settable {
		v.Set(nv)
	}
	return nil
}
func (ce ceArray) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	v = CastV(ce.t,v)
	n := ce.t.Len()
	for i:=0; i<n; i++ {
		e := ce.child.Write(w,v.Index(i))
		if e!=nil { return e }
	}
	return nil
}
type cePtr struct{
	child CodecElement
	t reflect.Type
}
func (ce cePtr) Read(r preciseio.PreciseReader,v reflect.Value) error {
	b,e := r.R.ReadByte()
	if e!=nil { return e }
	if b==0 {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	pv := reflect.New(ce.t.Elem())
	v.Set(pv)
	return ce.child.Read(r,pv.Elem())
}
func (ce cePtr) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	v = CastV(ce.t,v)
	if v.IsNil() { return w.W.WriteByte(0) }
	e := w.W.WriteByte(0xff)
	if e!=nil { return e }
	return ce.child.Write(w,v.Elem())
}

func StripawayPtr(i interface{}) CodecElement {
	t := reflect.TypeOf(i)
	ce := serializerFor(t.Elem())
	if ce==nil { return nil }
	return ceStripawayPtr{ce,t}
}
func StripawayPtrWith(i interface{}, ce CodecElement) CodecElement {
	t := reflect.TypeOf(i)
	if ce==nil { return nil }
	return ceStripawayPtr{ce,t}
}

type ceStripawayPtr struct{
	child CodecElement
	t reflect.Type
}
func (ce ceStripawayPtr) Read(r preciseio.PreciseReader,v reflect.Value) error {
	pv := reflect.New(ce.t.Elem())
	v.Set(pv)
	return ce.child.Read(r,pv.Elem())
}
func (ce ceStripawayPtr) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	v = CastV(ce.t,v)
	var ev reflect.Value
	if v.IsNil() { ev = reflect.Zero(ce.t.Elem()) } else { ev = v.Elem() }
	return ce.child.Write(w,ev)
}

func AddPtr(i interface{}) CodecElement {
	t := reflect.TypeOf(i)
	ce := serializerFor(t.Elem())
	if ce==nil { return nil }
	return ceAddPtr{ce,t}
}
func AddPtrWith(i interface{}, ce CodecElement) CodecElement {
	t := reflect.TypeOf(i)
	if ce==nil { return nil }
	return ceAddPtr{ce,t}
}

type ceAddPtr struct{
	child CodecElement
	t reflect.Type
}
func (ce ceAddPtr) Read(r preciseio.PreciseReader,v reflect.Value) error {
	ptr := reflect.New(ce.t).Elem()
	err := ce.child.Read(r,ptr)
	if !ptr.IsNil() { v.Set(ptr.Elem()) }
	return err
}
func (ce ceAddPtr) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	ptr := reflect.New(ce.t.Elem())
	ptr.Elem().Set(CastV(ce.t.Elem(),v))
	return ce.child.Write(w,ptr)
}

