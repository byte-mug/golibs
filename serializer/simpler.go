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

func Cast(t reflect.Type,i interface{}) (v reflect.Value) {
	v = reflect.ValueOf(i)
	if !v.IsValid() || v.Type()!=t { v = reflect.Zero(t) }
	return
}

func CastV(t reflect.Type,v reflect.Value) (reflect.Value) {
	if !v.IsValid() { return reflect.Zero(t) }
	if v.Type()!=t { v = reflect.ValueOf(v.Interface()) }
	if !v.IsValid() || v.Type()!=t { v = reflect.Zero(t) }
	return v
}
func GetInterface(v reflect.Value) interface{} {
	if !v.IsValid() { return nil }
	return v.Interface()
}
