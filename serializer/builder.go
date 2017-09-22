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

// Obtains the serializer for reflect.TypeOf(i)
func ForType(i interface{}) CodecElement {
	return serializerFor(reflect.TypeOf(i))
}
// Obtains the serializer for reflect.TypeOf(i).Elem()
func ForTypeP(i interface{}) CodecElement {
	return serializerFor(reflect.TypeOf(i).Elem())
}

func serializerFor(t reflect.Type) CodecElement {
	switch t.Kind() {
	case reflect.Slice:
		se := t.Elem()
		if se.Kind()==reflect.Uint8 { return ceBlob{} }
		r := serializerFor(se)
		if r==nil { return nil }
		return ceSlice{r,t}
	case reflect.String:
		return ceString{}
	case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
		return ceInt{}
	case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
		return ceUint{}
	case reflect.Map:
		mk := t.Key()
		me := t.Elem()
		mks := serializerFor(mk)
		if mks==nil { return nil }
		mes := serializerFor(me)
		if mes==nil { return nil }
		return ceMap{mks,mes,t}
	case reflect.Array:
		ae := t.Elem()
		aes := serializerFor(ae)
		return ceArray{aes,t}
	}
	return nil
}

// Obtains the serializer for reflect.TypeOf(i) and ce as the serializer for the contained elements.
func ForContainerWith(i interface{},ce CodecElement) CodecElement {
	return serializerForElem(reflect.TypeOf(i),1,ce)
}
// Obtains the serializer for reflect.TypeOf(i) and ce as the serializer for the contained elements (after depth).
//
// For example:
//	ForContainerWithDepth([][]*Bar{},2,fooBar)
// is equivalent to:
//	ForContainerWith([][]*Bar{},ForContainerWith([]*Bar{},fooBar)))
func ForContainerWithDepth(i interface{},depth int,ce CodecElement) CodecElement {
	return serializerForElem(reflect.TypeOf(i),depth,ce)
}

// Obtains the serializer for reflect.TypeOf(i).Elem() and ce as the serializer for the contained elements.
func ForContainerWithP(i interface{},ce CodecElement) CodecElement {
	return serializerForElem(reflect.TypeOf(i).Elem(),1,ce)
}
// Obtains the serializer for reflect.TypeOf(i).Elem() and ce as the serializer for the contained elements (after depth).
//
// For example:
//	ForContainerWithDepthP(new([2][2]*Bar),2,fooBar)
// is equivalent to:
//	ForContainerWithP(new([2][2]*Bar),ForContainerWithP(new([2]*Bar),fooBar))
func ForContainerWithDepthP(i interface{},depth int,ce CodecElement) CodecElement {
	return serializerForElem(reflect.TypeOf(i).Elem(),depth,ce)
}

func serializerForElem(t reflect.Type,depth int, contained CodecElement) CodecElement {
	if depth<1 { return contained }
	switch t.Kind() {
	case reflect.Slice:
		se := t.Elem()
		r := serializerForElem(se,depth-1,contained)
		if r==nil { return nil }
		return ceSlice{r,t}
	case reflect.Map:
		mk := t.Key()
		me := t.Elem()
		mks := serializerFor(mk)
		if mks==nil { return nil }
		mes := serializerForElem(me,depth-1,contained)
		if mes==nil { return nil }
		return ceMap{mks,mes,t}
	case reflect.Array:
		ae := t.Elem()
		aes := serializerForElem(ae,depth-1,contained)
		return ceArray{aes,t}
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
func (s *StructBuilder) FieldWith(name string,ce CodecElement) *StructBuilder{
	f,ok := s.t.Elem().FieldByName(name)
	if !ok { panic("No such field "+name) }
	s.fields = append(s.fields,strctField{ce,f.Index})
	return s
}
func (s *StructBuilder) FieldContainerWithDepth(name string,depth int,ce CodecElement) *StructBuilder{
	if ce==nil { panic("ce must not be <nil>") }
	f,ok := s.t.Elem().FieldByName(name)
	if !ok { panic("No such field "+name) }
	ser := serializerForElem(f.Type,depth,ce)
	if ser==nil { panic(fmt.Sprintf("Field %s: non-supported type: [depth=%d] %v",name,depth,f.Type)) }
	s.fields = append(s.fields,strctField{ser,f.Index})
	return s
}
func (s *StructBuilder) FieldContainerWith(name string,ce CodecElement) *StructBuilder{
	if ce==nil { panic("ce must not be <nil>") }
	f,ok := s.t.Elem().FieldByName(name)
	if !ok { panic("No such field "+name) }
	ser := serializerForElem(f.Type,1,ce)
	if ser==nil { panic(fmt.Sprintf("Field %s: non-supported type: [depth=1] %v",name,f.Type)) }
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
	v = CastV(s.t,v)
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

type typeSwitchItem struct{
	t reflect.Type
	ce CodecElement
}

// Creates a type-mapping for up to 255 different types.
// The type is indicated by the first byte.
type TypeSwitch struct{
	def byte
	oth map[byte]typeSwitchItem
}

// Starts a new type switching. A default key must be specified.
func Switch(def byte) *TypeSwitch {
	return &TypeSwitch{def,make(map[byte]typeSwitchItem)}
}
func (t *TypeSwitch) add(b byte, i typeSwitchItem) *TypeSwitch {
	_,hasByte1 := t.oth[b]
	hasByte2 := t.def==b
	if hasByte1||hasByte2 { panic(fmt.Sprintf("Name-Conflict: byte %d is already in use.",b)) }
	t.oth[b] = i
	return t
}
func (t *TypeSwitch) AddType(b byte,i interface{}) *TypeSwitch {
	tp := reflect.TypeOf(i)
	ce := serializerFor(tp)
	if ce==nil { panic(fmt.Sprintf("Switch %d: non-supported type: %v",b,tp)) }
	return t.add(b,typeSwitchItem{tp,ce})
}
func (t *TypeSwitch) AddTypeP(b byte,i interface{}) *TypeSwitch {
	tp := reflect.TypeOf(i).Elem()
	ce := serializerFor(tp)
	if ce==nil { panic(fmt.Sprintf("Switch %d: non-supported type: %v",b,tp)) }
	return t.add(b,typeSwitchItem{tp,ce})
}
func (t *TypeSwitch) AddTypeWith(b byte,i interface{}, ce CodecElement) *TypeSwitch {
	tp := reflect.TypeOf(i)
	if ce==nil { panic("ce must not be <nil>") }
	return t.add(b,typeSwitchItem{tp,ce})
}
func (t *TypeSwitch) AddTypeWithP(b byte,i interface{}, ce CodecElement) *TypeSwitch {
	tp := reflect.TypeOf(i).Elem()
	if ce==nil { panic("ce must not be <nil>") }
	return t.add(b,typeSwitchItem{tp,ce})
}

func (t *TypeSwitch) AddTypeContainerWithDepth(b byte,i interface{},depth int, ce CodecElement) *TypeSwitch {
	tp := reflect.TypeOf(i)
	if ce==nil { panic("ce must not be <nil>") }
	ser := serializerForElem(tp,depth,ce)
	if ser==nil { panic(fmt.Sprintf("Switch %d: non-supported type: [depth=%d] %v",b,depth,tp)) }
	return t.add(b,typeSwitchItem{tp,ser})
}
func (t *TypeSwitch) AddTypeContainerWithDepthP(b byte,i interface{},depth int, ce CodecElement) *TypeSwitch {
	tp := reflect.TypeOf(i).Elem()
	if ce==nil { panic("ce must not be <nil>") }
	ser := serializerForElem(tp,depth,ce)
	if ser==nil { panic(fmt.Sprintf("Switch %d: non-supported type: [depth=%d] %v",b,depth,tp)) }
	return t.add(b,typeSwitchItem{tp,ser})
}
func (t *TypeSwitch) AddTypeContainerWith(b byte,i interface{},ce CodecElement) *TypeSwitch {
	tp := reflect.TypeOf(i)
	if ce==nil { panic("ce must not be <nil>") }
	ser := serializerForElem(tp,1,ce)
	if ser==nil { panic(fmt.Sprintf("Switch %d: non-supported type: [depth=1] %v",b,tp)) }
	return t.add(b,typeSwitchItem{tp,ser})
}
func (t *TypeSwitch) AddTypeContainerWithP(b byte,i interface{},ce CodecElement) *TypeSwitch {
	tp := reflect.TypeOf(i).Elem()
	if ce==nil { panic("ce must not be <nil>") }
	ser := serializerForElem(tp,1,ce)
	if ser==nil { panic(fmt.Sprintf("Switch %d: non-supported type: [depth=1] %v",b,tp)) }
	return t.add(b,typeSwitchItem{tp,ser})
}

func (t *TypeSwitch) Read(r preciseio.PreciseReader,v reflect.Value) error {
	b,e := r.R.ReadByte()
	if e!=nil { return e }
	i,ok := t.oth[b]
	if !ok {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	return i.ce.Read(r,v)
}
func (t *TypeSwitch) Write(w *preciseio.PreciseWriter,v reflect.Value) error {
	dt := GetInterface(v)
	dv := reflect.ValueOf(dt)
	if !dv.IsValid() {
		return w.W.WriteByte(t.def)
	}
	dvt := dv.Type()
	for b,i := range t.oth {
		if dvt!=i.t { continue }
		e := w.W.WriteByte(b)
		if e!=nil { return e }
		return i.ce.Write(w,dv)
	}
	return w.W.WriteByte(t.def)
}

