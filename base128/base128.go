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

// A streaming interface to a Base128-Encoding
package base128

import "io"
import "bufio"

type Writer interface{
	io.WriteCloser
	Reset(w io.Writer)
}

type writer struct{
	w *bufio.Writer
	l int
	bits uint
	data uint64
}
func NewWriter(w io.Writer) Writer {
	return &writer{w:bufio.NewWriter(w)}
}

func (w *writer) llwrite(c byte) error {
	err := w.w.WriteByte(c|0x80)
	w.l++
	if w.l > 80 {
		w.w.WriteString("\r\n")
		w.l = 0
	}
	return err
}
func (w *writer) Reset(wr io.Writer) {
	w.l = 0
	w.bits = 0
	w.data = 0
	w.w.Reset(wr)
}
func (w *writer) Write(p []byte) (n int,e error) {
	bits := w.bits
	data := w.data
	n = len(p)
	for i,b := range p {
		data = (data<<8)|uint64(b)
		bits+=8
		bits-=7
		e = w.llwrite(byte(data>>bits))
		if e!=nil { n = i ; break }
		
		if bits>= 7 {
			bits-=7
			e = w.llwrite(byte(data>>bits))
			if e!=nil { n = i ; break }
		}
	}
	w.bits = bits
	w.data = data
	return
}
func (w *writer) Close() (e error) {
	bits := w.bits
	data := w.data
	data <<=6
	bits += 6
	for bits > 7 {
		bits-=7
		e = w.llwrite(byte(data>>bits))
		if e!=nil { break }
	}
	w.bits = 0
	w.data = 0
	if e==nil { e = w.w.Flush() }
	return
}

type Reader interface{
	io.Reader
	Reset(r io.Reader)
}

type reader struct{
	r *bufio.Reader
	bits uint
	data uint64
	e error
}
func NewReader(r io.Reader) Reader {
	return &reader{r:bufio.NewReader(r)}
}

func (r *reader) llread() (byte,error) {
	for {
		b,e := r.r.ReadByte()
		if e!=nil || (b&0x80)==0x80 { return b,e }
	}
	panic("unreachable")
}
func (r *reader) Reset(rd io.Reader) {
	r.bits = 0
	r.data = 0
	r.e = nil
	r.r.Reset(rd)
}
func (r *reader) Read(p []byte) (n int, err error) {
	if r.e!=nil { return 0,r.e }
	bits := r.bits
	data := r.data
	for i := range p {
		b,e := r.llread()
		if e!=nil { r.e = e ; err = e ; n = i ; break }
		data = (data<<7)|uint64(b&0x7f)
		bits+=7
		if bits<8 {
			b,e := r.llread()
			if e!=nil { r.e = e ; err = e ; n = i ; break }
			data = (data<<7)|uint64(b&0x7f)
			bits+=7
		}
		bits-=8
		p[i] = byte(data>>bits)
	}
	r.bits = bits
	r.data = data
	return
}

