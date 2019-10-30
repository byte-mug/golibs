/*
Copyright (c) 2017-2019 Simon Schmidt

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

package pentry

import "reflect"
import "encoding/binary"

func sizeof(v reflect.Value) int {
	switch v.Kind() {
	case reflect.Bool,reflect.Int8,reflect.Uint8: return 1
	case reflect.Int16,reflect.Uint16: return 2
	case reflect.Int32,reflect.Uint32: return 4
	case reflect.Int64,reflect.Uint64: return 8
	case reflect.Slice:
		if v.Type().Elem().Kind()!=reflect.Uint8 { break }
		return v.Len()+4
	case reflect.Array:
		if v.Len()>0 { return v.Len()*sizeof(v.Index(0)) }
	case reflect.Struct:
		size := 0
		for num,i := v.NumField(),0 ; i<num ; i++ {
			size += sizeof(v.Field(i))
		}
		return size
	}
	return 0
}

func fakeRead(v reflect.Value,buf []byte, bo binary.ByteOrder) (size int) {
	switch v.Kind() {
	case reflect.Bool,reflect.Int8,reflect.Uint8: return 1
	case reflect.Int16,reflect.Uint16: return 2
	case reflect.Int32,reflect.Uint32: return 4
	case reflect.Int64,reflect.Uint64: return 8
	case reflect.Slice:
		if v.Type().Elem().Kind()!=reflect.Uint8 { break }
		return int(bo.Uint32(buf))+4
	case reflect.Array:
		for i,n := 0,v.Len() ; i<n ; i++ {
			size += fakeRead(v.Index(i),buf[size:],bo)
		}
	case reflect.Struct:
		for i,n := 0,v.NumField() ; i<n ; i++ {
			size += fakeRead(v.Field(i),buf[size:],bo)
		}
	}
	return
}

func read(v reflect.Value,buf []byte, bo binary.ByteOrder) (size int) {
	if !v.CanSet() { return fakeRead(v,buf,bo) } // Fake-Read
	switch v.Kind() {
	case reflect.Bool:   v.SetBool(buf[0]!=0)                   ; return 1
	case reflect.Int8:   v.SetInt(int64(int8(buf[0])))          ; return 1
	case reflect.Int16:  v.SetInt(int64(int16(bo.Uint16(buf)))) ; return 2
	case reflect.Int32:  v.SetInt(int64(int32(bo.Uint32(buf)))) ; return 4
	case reflect.Int64:  v.SetInt(int64(bo.Uint64(buf)))        ; return 8
	
	case reflect.Uint8:  v.SetUint(uint64(buf[0]))         ; return 1
	case reflect.Uint16: v.SetUint(uint64(bo.Uint16(buf))) ; return 2
	case reflect.Uint32: v.SetUint(uint64(bo.Uint32(buf))) ; return 4
	case reflect.Uint64: v.SetUint(bo.Uint64(buf))         ; return 8
	case reflect.Slice:
		if v.Type().Elem().Kind()!=reflect.Uint8 { break }
		bufl := bo.Uint32(buf)
		v.SetBytes(buf[4:][:bufl])
		return int(bufl)+4
	case reflect.Array:
		for i,n := 0,v.Len() ; i<n ; i++ {
			size += read(v.Index(i),buf[size:],bo)
		}
	case reflect.Struct:
		for i,n := 0,v.NumField() ; i<n ; i++ {
			size += read(v.Field(i),buf[size:],bo)
		}
	}
	return
}
func write(v reflect.Value,buf []byte, bo binary.ByteOrder) (size int) {
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() { buf[0]=0xff } else { buf[0]=0 }
		return 1
	case reflect.Int8: buf[0] = byte(int8(v.Int()))       ; return 1
	case reflect.Int16: bo.PutUint16(buf,uint16(v.Int())) ; return 2
	case reflect.Int32: bo.PutUint32(buf,uint32(v.Int())) ; return 4
	case reflect.Int64: bo.PutUint64(buf,uint64(v.Int())) ; return 8
	
	case reflect.Uint8: buf[0] = byte(v.Uint())             ; return 1
	case reflect.Uint16: bo.PutUint16(buf,uint16(v.Uint())) ; return 2
	case reflect.Uint32: bo.PutUint32(buf,uint32(v.Uint())) ; return 4
	case reflect.Uint64: bo.PutUint64(buf,v.Uint())         ; return 8
	case reflect.Slice:
		if v.Type().Elem().Kind()!=reflect.Uint8 { break }
		bo.PutUint32(buf,uint32(v.Len()))
		copy(buf[4:],v.Bytes())
		return v.Len()+4
	case reflect.Array:
		for i,n := 0,v.Len() ; i<n ; i++ {
			size += write(v.Index(i),buf[size:],bo)
		}
	case reflect.Struct:
		for i,n := 0,v.NumField() ; i<n ; i++ {
			size += write(v.Field(i),buf[size:],bo)
		}
	}
	return
}

func Sizeof(i interface{}) int {
	return sizeof(reflect.Indirect(reflect.ValueOf(i)))
}
func BufferSizeof(i interface{},buf []byte, bo binary.ByteOrder) int {
	return fakeRead(reflect.Indirect(reflect.ValueOf(i)),buf,bo)
}
func ReadSize(i interface{},buf []byte, bo binary.ByteOrder) {
	read(reflect.Indirect(reflect.ValueOf(i)),buf,bo)
}
func Read(i interface{},buf []byte, bo binary.ByteOrder) {
	read(reflect.Indirect(reflect.ValueOf(i)),buf,bo)
}
func Write(i interface{},buf []byte, bo binary.ByteOrder) {
	write(reflect.Indirect(reflect.ValueOf(i)),buf,bo)
}

