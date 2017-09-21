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

import "fmt"
import "reflect"
import "github.com/byte-mug/golibs/preciseio"

func serializerFor(t reflect.Type) CodecElement {
	switch t.Kind() {
	case reflect.Slice:
		se := t.Elem()
		if se.Kind()==reflect.Uint8 { return ceBlob{} }
		r := serializerFor(se)
		if r==nil { return nil }
		return ceSlice{r,se}
	case reflect.String:
		return ceString{}
	case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
		return ceInt{}
	case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
		return ceUint{}
	}
	return nil
}

type strctField struct{
	ce   CodecElement
	idxs []int
}
type StructBuilder struct{
	t reflect.Type
	fields []strctField
}
func With(i interface{}) *StructBuilder {
	ti := reflect.TypeOf(i)
	if ti.Kind()!=reflect.Ptr || ti.Elem().Kind()!=reflect.Struct { panic(fmt.Sprintf("Required *struct{}, but got %v",ti)) }
	return &StructBuilder{ti,nil}
}
func (s *StructBuilder) Field(name string) *StructBuilder{
	f,ok := s.t.Elem().FieldByName(name)
	if !ok { panic("No such field "+name) }
	ser := serializerFor(f.Type)
	if ser==nil { panic(fmt.Sprintf("Field %s: non-supported type: %v",name,f.Type)) }
	s.fields = append(s.fields,strctField{ser,f.Index})
	return s
}
func (s *StructBuilder) Read(r preciseio.PreciseReader,v reflect.Value) error {
	b,e := r.R.ReadByte()
	if e!=nil { return e }
	if b==0 {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	pv := reflect.New(s.t.Elem())
	defer v.Set(pv)
	ev := pv.Elem()
	for _,field := range s.fields {
		e = field.ce.Read(r,ev.FieldByIndex(field.idxs))
		if e!=nil { return e }
	}
	return nil
}
func (s *StructBuilder) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	v = Cast(s.t,v.Interface())
	if v.IsNil() {
		return w.W.WriteByte(0)
	}
	ev := v.Elem()
	e := w.W.WriteByte(0xff)
	if e!=nil { return e }
	for _,field := range s.fields {
		e = field.ce.Write(w,ev.FieldByIndex(field.idxs))
		if e!=nil { return e }
	}
	return nil
}


