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
import "net"
import "sync"
import "bufio"
import "time"

type Server struct{
	NewMessage func() *SendMessage
	Handler    func(r *SendMessage,c *ServerConn)
}

type timeoutReader struct{
	r net.Conn
	e error
}
func (r *timeoutReader) Read(p []byte) (i int,e error) {
	if r.e!=nil { return 0,r.e }
	r.r.SetReadDeadline(coarseNow.Add(time.Second*5))
	i,e = r.r.Read(p)
	if r.e==nil && e!=nil { r.e = e }
	return
}

type ServerConn struct{
	nc net.Conn
	br *timeoutReader
	tr *textproto.Reader
	tw *textproto.Writer
	
	wrl sync.Mutex
	
	*Server
}
func (c *ServerConn) worker(){
	defer c.nc.Close()
	for {
		msg := c.NewMessage()
		err := DecodeMessage(c.tr,msg)
		if err==nil {
			go c.Handler(msg,c)
		}
		if c.br.e!=nil { break }
	}
}
func (c *ServerConn) WriteMessage(msg *SendMessage) error {
	c.wrl.Lock(); defer c.wrl.Unlock()
	return EncodeMessage(c.tw,msg)
}

func (s *Server) Handle(c net.Conn) {
	cc := new(ServerConn)
	cc.nc = c
	cc.br = &timeoutReader{r:c}
	cc.tr = textproto.NewReader(bufio.NewReader(cc.br))
	cc.tw = textproto.NewWriter(bufio.NewWriter(c))
	cc.Server = s
	cc.worker()
}

