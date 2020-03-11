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
import "unsafe"

type ro_t uint8
const (
	ro_none ro_t = iota
	ro_ptr
	ro_bytes
	ro_struct
)

type reflectType struct{
	temp reflect.Type
	size int
	ro_t ro_t
	ftyp []reflectType
}

type reflectObject struct{
	*reflectType
	pset *unsafe.Pointer
	bset *[]byte
	fobj []reflectObject
}

func (r *reflectObject) init1(vl reflect.Value) {
	if ftl := len(r.ftyp); ftl>0 {
		r.fobj = make([]reflectObject,ftl)
	} else {
		r.fobj = nil
	}
	switch r.ro_t {
	case ro_ptr:
		/*
		 * The value 'vl' is something like *uint32,
		 * so 'vl.Addr()' is something like **uint32
		 * which, in turn, is straight a pointer to a pointer,
		 * which validly is *unsafe.Pointer
		 */
		r.pset = (*unsafe.Pointer)((unsafe.Pointer)(vl.Addr().Pointer()))
	case ro_bytes:
		r.bset = vl.Addr().Interface().(*[]byte)
	}
	for i := range r.fobj {
		r.fobj[i].reflectType = &r.ftyp[i]
		r.fobj[i].init1(vl.Field(i))
	}
}
func (r *reflectObject) setBytes(buf []byte) {
	if len(buf)!=r.size { panic("invalid size") }
	switch r.ro_t {
	case ro_ptr: *r.pset = unsafe.Pointer(&buf[0])
	case ro_bytes: *r.bset = buf[:r.size:r.size]
	case ro_struct:
		for i := range r.fobj {
			sz := r.fobj[i].size
			r.fobj[i].setBytes(buf[:sz])
			buf = buf[sz:]
		}
	}
}

type topLevelObject struct{
	reflectObject
	val interface{}
	rval reflect.Value
}
func (r *topLevelObject) init2() {
	var targ,self reflect.Value
	switch r.ro_t {
	case ro_ptr, ro_bytes:
		self = reflect.New(r.temp)
		targ = self.Elem()
	case ro_struct:
		self = reflect.New(r.temp)
		targ = self.Elem()
		r.val = self.Interface()
	}
	r.rval = targ
	r.reflectObject.init1(targ)
}
func (r *topLevelObject) SetBytes(buf []byte) {
	r.setBytes(buf)
}
func (r *topLevelObject) Value() interface{} {
	if r.val==nil { return r.rval.Interface() }
	return r.val
}
var _ Instance = (*topLevelObject)(nil)

type iAccessType struct{
	reflectType
}
func (at *iAccessType) New() Instance {
	tlo := new(topLevelObject)
	tlo.reflectType = &at.reflectType
	tlo.init2()
	return tlo
}
func (at *iAccessType) Len() int {
	return at.size
}
var _ AccessType = (*iAccessType)(nil)

