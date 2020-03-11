/*
Copyright (c) 2020 Simon Schmidt

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


package memstruct

import "reflect"
import "fmt"
import "strconv"

func (t *reflectType) buildType(tp reflect.Type) {
	switch tp.Kind() {
	case reflect.Struct,reflect.Ptr,reflect.Slice: /* smile */
	default: tp = reflect.PtrTo(tp)
	}
	t.build(reflect.StructField{Type:tp})
}
func (t *reflectType) build(sf reflect.StructField) {
	t.temp = sf.Type
	switch t.temp.Kind() {
	case reflect.Struct: t.buildStruct()
	case reflect.Ptr: t.buildPointer()
	case reflect.Slice:
		if t.temp.Elem().Kind()!=reflect.Uint8 {
			panic(fmt.Sprintf("unsupported type: %v",t.temp))
		}
		t.ro_t = ro_bytes
		t.setSize(sf)
	default: panic(fmt.Sprintf("unsupported type: %v",t.temp))
	}
}
func (t *reflectType) setSize(sf reflect.StructField) {
	ri,_ := strconv.ParseInt(sf.Tag.Get("bytes"),0,64)
	i := int(ri)
	if i<0 { i = 0 }
	t.size = i
}
func (t *reflectType) buildStruct() {
	t.ftyp = make([]reflectType,t.temp.NumField())
	t.ro_t = ro_struct
	t.size = 0
	for i := range t.ftyp {
		t.ftyp[i].build(t.temp.Field(i))
		t.size += t.ftyp[i].size
	}
}
func (t *reflectType) buildPointer() {
	t.ro_t = ro_ptr
	switch t.temp.Elem().Kind() {
	case reflect.Int8,reflect.Uint8: t.size = 1
	case reflect.Int16,reflect.Uint16: t.size = 2
	case reflect.Int32,reflect.Uint32: t.size = 4
	case reflect.Int64,reflect.Uint64: t.size = 8
	default: panic(fmt.Sprintf("unsupported pointer type: %v",t.temp))
	}
}

