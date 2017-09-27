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

package quickdump

import "github.com/byte-mug/golibs/preciseio"
import "reflect"
import "fmt"
import "math"

const ourTag = "quickdump"

func decodeTag2(sf reflect.StructField) (string,string) {
	s := sf.Tag.Get(ourTag)
	for i,b := range []byte(s) {
		if b==',' { return s[:i],s[i+1:] }
	}
	return s,""
}
func iterateString(s string) (string,string) {
	for i,b := range []byte(s) {
		if b==',' { return s[:i],s[i+1:] }
	}
	return s,""
}

func getString(s []string, i int) string {
	if len(s)<=i { return "" }
	return s[i]
}
func isNULL(v reflect.Value) bool {
	return v.Kind()==reflect.Ptr && v.IsNil()
}

func nullableIsNULL(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr: return v.IsNil()
	case reflect.Struct: return !v.Field(0).Bool()
	}
	return false
}

//func alibi(i ...interface{}) {}

func stripNullable(isR, isW bool, v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Ptr: return elem(isR,isW,v)
	case reflect.Struct:
		v.Field(0).SetBool(true)
		return v.Field(1)
	}
	panic(fmt.Sprint("This type is not nullable: ",v.Type()," of kind ",v.Kind()))
}

func elem(isR, isW bool, v reflect.Value) reflect.Value {
	if isR {
		nv := reflect.New(v.Type().Elem())
		v.Set(nv)
		return nv.Elem()
	}
	if isW {
		if v.IsNil() { return reflect.Zero(v.Type().Elem()) }
		return v.Elem()
	}
	return reflect.Zero(v.Type().Elem())
}

func length(t reflect.Type,i,n int) (l int) {
	l = 1
	i++
	for ; i<n ; i++ {
		tag,_ := decodeTag2(t.Field(i))
		if tag!="more" { break }
		l++
	}
	return
}
func findNonNULL(v reflect.Value,i,n int) (l int) {
	l = 0
	for ; i<n ; i++ {
		if !nullableIsNULL(v.Field(i)) { return }
		l++
	}
	l = -1
	return
}

func vperformStruct(isR, isW bool, r preciseio.PreciseReader, w *preciseio.PreciseWriter, v reflect.Value) (err error) {
	t := v.Type()
	i := 0
	n := t.NumField()
	incr := 0
	
	//wasNULL := true
	for ; i<n ; i+=incr {
		tag,more := decodeTag2(t.Field(i))
		fv := v.Field(i)
		incr = 1
		switch tag {
		case "tag":
			incr = length(t,i,n)
			tag,more = iterateString(more)
		
		case "more": // Blind tags
			tag,more = iterateString(more)
		}
		
		idx := 0
		if incr>1 {
			if isR {
				idx,err = r.ReadListLength()
				if err!=nil { return }
			}
			if isW {
				idx = findNonNULL(v,i,i+incr)
				if idx<0 { idx=incr }
				err = w.WriteListLength(idx)
				if err!=nil { return }
			}
			if idx<incr {
				fv = v.Field(i+idx)
				tag,more = decodeTag2(t.Field(i+idx))
				tag,more = iterateString(more) // strip first tag.
			} else {
				continue
			}
		}
		
		perform := true
		for ;; tag,more = iterateString(more) {
			switch tag {
			case "strip":
				fv = stripNullable(isR,isW,fv)
				continue
			case "nullable":
				if isR {
					b,e := r.R.ReadByte()
					if e!=nil { return e }
					perform = b!=0
				}
				if isW {
					perform = !nullableIsNULL(fv)
					b := byte(0)
					if perform { b = 0xff }
					err = w.W.WriteByte(b)
					if err!=nil { return }
				}
				if perform {
					fv = stripNullable(isR,isW,fv)
					continue
				}
			}
			break
		}
		
		if !perform { continue }
		e := vperform(isR,isW,r,w,fv)
		//wasNULL = isNULL(fv)
		
		if e!=nil { return e }
	}
	return nil
}

func Marshal(w *preciseio.PreciseWriter,i interface{}) error {
	return vperform(false,true,preciseio.PreciseReader{},w,reflect.ValueOf(i).Elem())
}

func Unmarshal(r preciseio.PreciseReader,i interface{}) error {
	v := reflect.ValueOf(i).Elem()
	v.Set(reflect.Zero(v.Type()))
	return vperform(true,false,r,nil,v)
}

func vperform(isR, isW bool, r preciseio.PreciseReader, w *preciseio.PreciseWriter, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Bool:
		if isR {
			b,e := r.R.ReadByte()
			v.SetBool(b!=0)
			return e
		}
		if isW {
			b := byte(0)
			if v.Bool() { b=0xff }
			return w.W.WriteByte(b)
		}
		return nil
	case reflect.Int,reflect.Int16,reflect.Int32,reflect.Int64:
		if isR {
			i,e := r.ReadVarint()
			v.SetInt(i)
			return e
		}
		if isW {
			i := v.Int()
			return w.WriteVarint(i)
		}
		return nil
	case reflect.Int8:
		if isR {
			b,e := r.R.ReadByte()
			i := int8(b)
			v.SetInt(int64(i))
			return e
		}
		if isW {
			i := int8(v.Int())
			return w.W.WriteByte(byte(i))
		}
		return nil
	case reflect.Uint,reflect.Uint16,reflect.Uint32,reflect.Uint64,reflect.Uintptr:
		if isR {
			i,e := r.ReadUvarint()
			v.SetUint(i)
			return e
		}
		if isW {
			i := v.Uint()
			return w.WriteUvarint(i)
		}
		return nil
	case reflect.Uint8:
		if isR {
			b,e := r.R.ReadByte()
			v.SetUint(uint64(b))
			return e
		}
		if isW {
			i := v.Uint()
			return w.W.WriteByte(byte(i))
		}
		return nil
	case reflect.Complex64:
		if isR {
			rp,e := r.ReadUvarint()
			if e!=nil { return e }
			ip,e := r.ReadUvarint()
			if e!=nil { return e }
			rv := float64(math.Float32frombits(uint32(rp)))
			iv := float64(math.Float32frombits(uint32(ip)))
			v.SetComplex(complex(rv,iv))
			return nil
		}
		if isW {
			c := v.Complex()
			rp := uint64(math.Float32bits(float32(real(c))))
			ip := uint64(math.Float32bits(float32(imag(c))))
			if e := w.WriteUvarint(rp) ; e!=nil { return e }
			return w.WriteUvarint(ip)
		}
		return nil
	case reflect.Complex128:
		if isR {
			rp,e := r.ReadUvarint()
			if e!=nil { return e }
			ip,e := r.ReadUvarint()
			if e!=nil { return e }
			rv := math.Float64frombits(rp)
			iv := math.Float64frombits(ip)
			v.SetComplex(complex(rv,iv))
			return nil
		}
		if isW {
			c := v.Complex()
			rp := math.Float64bits(real(c))
			ip := math.Float64bits(imag(c))
			if e := w.WriteUvarint(rp) ; e!=nil { return e }
			return w.WriteUvarint(ip)
		}
		return nil
	case reflect.Float32:
		if isR {
			i,e := r.ReadUvarint()
			if e!=nil { return e }
			v.SetFloat(float64(math.Float32frombits(uint32(i))))
			return nil
		}
		if isW {
			f := v.Float()
			i := uint64(math.Float32bits(float32(f)))
			return w.WriteUvarint(i)
		}
		return nil
	case reflect.Float64:
		if isR {
			i,e := r.ReadUvarint()
			if e!=nil { return e }
			v.SetFloat(math.Float64frombits(i))
			return nil
		}
		if isW {
			f := v.Float()
			i := math.Float64bits(f)
			return w.WriteUvarint(i)
		}
		return nil
	case reflect.Slice:
		if v.Type().Elem().Kind()==reflect.Uint8 {
			if isR {
				blob,e := r.ReadBlob()
				v.SetBytes(blob)
				return e
			}
			if isW {
				blob := v.Bytes()
				return w.WriteBlob(blob)
			}
			return nil
		}
		if isR {
			n,e := r.ReadListLength()
			if e!=nil { return e }
			nv := reflect.MakeSlice(v.Type(),n,n)
			v.Set(nv)
			for i:=0 ; i<n ; i++ {
				e = vperform(isR, isW, r, w, v.Index(i))
				if e!=nil { return e }
			}
			return nil
		}
		if isW {
			n := v.Len()
			e := w.WriteListLength(n)
			if e!=nil { return e }
			for i:=0 ; i<n ; i++ {
				e = vperform(isR, isW, r, w, v.Index(i))
				if e!=nil { return e }
			}
			return nil
		}
		return nil
	case reflect.Array:
		n := v.Type().Len()
		for i:=0 ; i<n ; i++ {
			e := vperform(isR, isW, r, w, v.Index(i))
			if e!=nil { return e }
		}
		return nil
	case reflect.Struct:
		return vperformStruct(isR, isW, r, w, v)
	case reflect.String:
		{
			if isR {
				blob,e := r.ReadBlob()
				v.SetString(string(blob))
				return e
			}
			if isW {
				blob := []byte(v.String())
				return w.WriteBlob(blob)
			}
			return nil
		}
	case reflect.Ptr:
		if isR {
			b,e := r.R.ReadByte()
			if e != nil { return e }
			if b==0 {
				v.Set(reflect.Zero(v.Type()))
				return nil
			}
			nv := reflect.New(v.Type().Elem())
			v.Set(nv)
			return vperform(isR, isW, r, w, nv.Elem())
		}
		if isW {
			if v.IsNil() { return w.W.WriteByte(0) }
			e := w.W.WriteByte(0xff)
			if e!=nil { return e }
			return vperform(isR, isW, r, w, v.Elem())
		}
		return nil
	}
	panic(fmt.Sprint("unsupported type: ",v.Type()," of kind ",v.Kind()))
}



