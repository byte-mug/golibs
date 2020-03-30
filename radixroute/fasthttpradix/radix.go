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


package fasthttpradix


import (
	"github.com/valyala/fasthttp"
	rr "github.com/byte-mug/golibs/radixroute"
	"strings"
)


func NormPath(str string) string {
	return "/"+strings.Trim(str,"/")+"/"
}

func InvalidPath(ctx *fasthttp.RequestCtx) {
	ctx.SetBody([]byte("404 Not Found\n"))
	ctx.SetStatusCode(fasthttp.StatusNotFound)
}

type handling struct{
	handle fasthttp.RequestHandler
}
type Router struct{
	routes map[string]*rr.Tree
}
func (r *Router) getOrCreate(m string) *rr.Tree {
	t := r.routes[m]
	if t!=nil { return t }
	if r.routes==nil { r.routes = make(map[string]*rr.Tree) }
	r.routes[m] = rr.New()
	return r.routes[m]
}
func (r *Router) Handle(method string,path string,handler fasthttp.RequestHandler) {
	if handler==nil { panic("handler must not be nil") }
	h := &handling{handler}
	cpath := []byte(NormPath(path))
	bpath := cpath[:len(cpath)-1]
	
	if method=="" {
		method = "DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT"
	}
	for _,m := range strings.Split(method,",") {
		rrt := r.getOrCreate(m)
		rrt.InsertRoute(bpath,h)
		rrt.InsertRoute(cpath,h)
	}
}
func (r *Router) RequestHandler(ctx *fasthttp.RequestCtx) {
	t := r.routes[string(ctx.Method())]
	if t==nil {
		InvalidPath(ctx)
		return
	}
	i,_ := t.Get(ctx.Path(),ctx.SetUserValue)
	h,ok := i.(*handling)
	if !ok {
		InvalidPath(ctx)
		return
	}
	h.handle(ctx)
}

