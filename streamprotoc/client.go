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

import "net"
import "net/textproto"
import "io"
import "bufio"
import "encoding/binary"
import "sync"
import "time"
import "math/rand"
import "errors"

var EAgain = errors.New("Ooops, try again")
var ETimeout = errors.New("Deadline timed out")

var randSource = rand.New(rand.NewSource(time.Now().Unix()^int64(time.Now().Nanosecond())))

const MZ1 = 12444583056938131134
const MZ2 = 7945676850369339463

type SendMessage struct{
	Inner Message
	ConnID uint64
	ReqID  uint64
}
func (s *SendMessage) WriteMsg(w *bufio.Writer) error {
	head := [4]uint64{MZ1,s.ConnID,s.ReqID,MZ2}
	err := binary.Write(w, binary.BigEndian,&head)
	if err!=nil { return err }
	return s.Inner.WriteMsg(w)
}
func (s *SendMessage) ReadMsg(r *bufio.Reader) error {
	var head [4]uint64
	err := binary.Read(r, binary.BigEndian,&head)
	if err!=nil { return err }
	s.ConnID = head[1]
	s.ReqID  = head[2]
	return s.Inner.ReadMsg(r)
}

type waitMessage struct{
	msg Message
	wg  sync.WaitGroup
	dl  time.Time
	e   error
}

type isBrokenReader struct{
	r io.Reader
	e error
}
func (r *isBrokenReader) Read(p []byte) (i int,e error) {
	i,e = r.r.Read(p)
	if r.e==nil && e!=nil { r.e = e }
	return
}

type Client struct{
	nc net.Conn
	br *isBrokenReader
	tr *textproto.Reader
	tw *textproto.Writer
	
	wrl sync.Mutex
	
	lock sync.Mutex
	
	id uint64
	mp map[uint64]*waitMessage
}
// Don't call this method!
func (c *Client) WriteMsg(w *bufio.Writer) error { panic("dont call me") }
// Don't call this method directly!
func (c *Client) ReadMsg(r *bufio.Reader) error {
	var head [4]uint64
	err := binary.Read(r, binary.BigEndian,&head)
	if err!=nil { return err }
	if head[0]!=MZ1 || head[3]!=MZ2 { return nil }
	ConnID := head[1]
	ReqID  := head[2]
	if ConnID!=c.id { return nil }
	
	c.lock.Lock()
	wm := c.mp[ReqID]
	delete(c.mp,ReqID)
	c.lock.Unlock()
	
	if wm==nil { return nil }
	
	wm.wg.Done()
	
	wm.e = wm.msg.ReadMsg(r)
	return wm.e
}
func (c *Client) worker(){
	defer c.nc.Close()
	for {
		DecodeMessage(c.tr,c)
		if c.br.e!=nil { break }
	}
}
func (c *Client) unFreezer(){
	ticker := time.NewTicker(time.Millisecond*100)
	defer ticker.Stop()
	for {
		t := <- ticker.C
		c.lock.Lock()
		for id,wm := range c.mp {
			if wm.dl.IsZero() { continue }
			if wm.dl.After(t) { continue }
			delete(c.mp,id)
			wm.e = ETimeout
			wm.wg.Done()
		}
		c.lock.Unlock()
		if c.br.e!=nil { break }
	}
	for _,wm := range c.mp {
		wm.e = c.br.e
		wm.wg.Done()
	}
}
func (c *Client) DoDeadline(req, resp Message, dl time.Time) error {
	for i := 0 ; i<16 ; i++ {
		err := c.innerDoDeadline(req,resp,dl)
		if err!=EAgain { return err }
		if err==nil { return nil }
		if !time.Now().Before(dl) { return ETimeout }
	}
	return EAgain
}
func (c *Client) innerDoDeadline(req, resp Message, dl time.Time) error {
	obj := &waitMessage{msg:resp,dl:dl}
	obj.wg.Add(1)
	ReqID := randSource.Uint64()
	c.lock.Lock()
	if c.mp[ReqID]!=nil {
		c.lock.Unlock()
		return EAgain
	}
	c.mp[ReqID] = obj
	c.lock.Unlock()
	
	c.wrl.Lock(); defer c.wrl.Unlock()
	err := EncodeMessage(c.tw,&SendMessage{req,c.id,ReqID})
	obj.wg.Wait()
	if err==nil { err = obj.e }
	return err
}
func NewClient(c net.Conn) *Client {
	cc := new(Client)
	cc.nc = c
	cc.br = &isBrokenReader{r:c}
	cc.tr = textproto.NewReader(bufio.NewReader(cc.br))
	cc.tw = textproto.NewWriter(bufio.NewWriter(c))
	cc.id = randSource.Uint64()
	cc.mp = make(map[uint64]*waitMessage)
	go cc.worker()
	go cc.unFreezer()
	return cc
}

