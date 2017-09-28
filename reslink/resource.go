/*
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


package reslink

import "sync"
import "sync/atomic"
import "container/list"

type Resource interface{
	Open() error
	Close() error
}

type ResourceElement struct{
	res Resource
	refc *int64
	refx bool
	refy bool
	elem *list.Element
	mutex sync.Mutex
}
func (r *ResourceElement) Incr() {
	atomic.AddInt64(r.refc, 1)
}
func (r *ResourceElement) Decr() {
	i := atomic.AddInt64(r.refc, -1)
	if i>0 { return }
	r.mutex.Lock(); defer r.mutex.Unlock()
	if !r.refy { return }
	r.refx = false
	r.refy = false
	r.res.Close()
}
func (r *ResourceElement) open() error {
	if r.refx { return nil }
	r.mutex.Lock(); defer r.mutex.Unlock()
	if r.refy { return nil }
	err := r.res.Open()
	if err!=nil { return nil }
	r.refx = true
	r.refy = true
	return nil
}
func NewResourceElement(res Resource) *ResourceElement {
	return &ResourceElement{
		res: res,
		refc: new(int64),
		refx: false,
		refy: false,
		elem: nil,
	}
}

type ResourceList struct{
	max  int
	list *list.List
	mutex sync.Mutex
}
func (r *ResourceList) Open(re *ResourceElement) error {
	r.mutex.Lock(); defer r.mutex.Unlock()
	err := re.open()
	if err!=nil { return err }
	if re.elem!=nil {
		r.list.MoveToFront(re.elem)
		return nil
	}
	re.Incr()
	re.elem = r.list.PushFront(re)
	
	i := r.list.Len()
	for ; i>r.max ; i-- {
		b := r.list.Back()
		b.Value.(*ResourceElement).Decr()
		r.list.Remove(b)
	}
	return nil
}
func NewResourceList(max int) *ResourceList{
	return &ResourceList{max: max, list: list.New() }
}


