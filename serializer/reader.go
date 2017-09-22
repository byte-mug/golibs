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
import "sync"

var tpAny = reflect.TypeOf(new(interface{})).Elem()
var pool_Values = sync.Pool{ New: func() interface{} { return reflect.New(tpAny).Elem() } }
func pool_Values_Put(v reflect.Value) {
	v.Set(reflect.Zero(tpAny))
	pool_Values.Put(v)
}

func Deserialize(ce CodecElement, r preciseio.PreciseReader) (interface{},error) {
	v := pool_Values.Get().(reflect.Value)
	defer pool_Values_Put(v)
	e := ce.Read(r,v)
	return v.Interface(),e
}
func Serialize(ce CodecElement, w *preciseio.PreciseWriter, i interface{}) error {
	return ce.Write(w,reflect.ValueOf(i))
}


