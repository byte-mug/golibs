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


package streamprotoc

import "net/textproto"
import "io/ioutil"
import "io"
import "bufio"
import "sync"
import "bytes"

import "github.com/byte-mug/golibs/base128"

var empty = bytes.NewReader([]byte{})

var bufr = sync.Pool{ New: func() interface{} { return bufio.NewReader(empty) } }
var bufw = sync.Pool{ New: func() interface{} { return bufio.NewWriter(ioutil.Discard) } }

var decs = sync.Pool{ New: func() interface{} { return base128.NewReader(empty) } }
var encs = sync.Pool{ New: func() interface{} { return base128.NewWriter(ioutil.Discard) } }

type Message interface{
	WriteMsg(w *bufio.Writer) error
	ReadMsg(r *bufio.Reader) error
}


func DecodeMessage(tr *textproto.Reader, msg Message) error {
	rdr := tr.DotReader()
	defer io.Copy(ioutil.Discard,rdr)
	r := decs.Get().(base128.Reader)
	defer decs.Put(r)
	r.Reset(rdr)
	br := bufr.Get().(*bufio.Reader)
	defer bufr.Put(br)
	br.Reset(r)
	return msg.ReadMsg(br)
}

func EncodeMessage(tw *textproto.Writer, msg Message) error {
	wrt := tw.DotWriter()
	defer wrt.Close()
	w := encs.Get().(base128.Writer)
	defer encs.Put(w)
	w.Reset(wrt)
	bw := bufw.Get().(*bufio.Writer)
	defer bufw.Put(bw)
	bw.Reset(w)
	err := msg.WriteMsg(bw)
	bw.Flush()
	w.Close()
	return err
}

