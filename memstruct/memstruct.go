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

/*
Implements orthogonal structures.

	// total 22 bytes.
	type MyStructure struct {
		Field1 *uint64 // 8 bytes
		Field2 *int16  // 2 bytes
		MyData []byte `bytes:"12"` // 12 bytes.
	}
	
	atype := memstruct.MakeAccessType(new(MyStructure))
	inst := atype.New()
	inst.SetBytes(make([]byte,22))
	ms := inst.Value().(*MyStructure)
	
	fmt.Println(*ms.Field1)
	fmt.Println(*ms.Field2)
	fmt.Println(ms.MyData)
*/
package memstruct

import "reflect"

func MakeAccessType(i interface{}) AccessType {
	tp := reflect.Indirect(reflect.ValueOf(i)).Type()
	at := new(iAccessType)
	at.buildType(tp)
	return at
}

type AccessType interface{
	New() Instance
	Len() int
}
type Instance interface{
	Value() interface{}
	SetBytes([]byte)
}

