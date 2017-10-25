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


package flatfile

import "encoding/binary"
import "io"
import "errors"
import "sync"
import "sort"
import "github.com/byte-mug/golibs/buffer"
import "fmt"

const (
	MaxRawSize     = (1<<24)      // 16 MiB
	MaxRecordSize  = MaxRawSize-4 // 16 MiB - 4 Byte
)

var EBadRecord = errors.New("Bad Record")
var ERecordTooLong = errors.New("Record Too Long")

var bE = binary.BigEndian

var ignoreMe = fmt.Errorf("hello")


/*
Scans a flat-file until the end, and then it returns its length.

It will always return an error, usually EBadRecord or io.EOF, depending on the consistency of the file.
If the file was corrupted, it will read until the last usable record.
*/
func ScanFlatFile(r io.ReaderAt) (recs int,off int64,err error) {
	var buf [4]byte
	for {
		_,err = r.ReadAt(buf[:],off)
		if err!=nil { return }
		recl := bE.Uint32(buf[:])
		if recl>MaxRawSize || recl<4 { err = EBadRecord ; return }
		off += int64(recl)
		recs++
	}
	return
}

func scanFlatFileAt(r io.ReaderAt,recs0 int, offset0 int64) (recs int,off int64,err error) {
	var buf [4]byte
	recs = recs0
	off  = offset0
	for {
		_,err = r.ReadAt(buf[:],off)
		if err!=nil { return }
		recl := bE.Uint32(buf[:])
		if recl>MaxRawSize || recl<4 { err = EBadRecord ; return }
		off += int64(recl)
		recs++
	}
	return
}

type ReaderWriter interface{
	io.ReaderAt
	io.WriterAt
}

type FlatFileWriter struct{
	w io.WriterAt
	nextRec int
	length  int64
	buf     [4]byte
}
/*
Initializes the FFWriter, assuming, that dest is empty.
*/
func (w *FlatFileWriter) InitNew(dest io.WriterAt) {
	w.w = dest
	w.nextRec = 0
	w.length  = 0
}

// Danger-Zone. Use this function, if and only if, you know what you are doing.
//
// This function initializes the FFWriter using a given number of records, and a given length.
// If the number of recs is incorrect, the returned ids will be wrong.
// If the length is incorrect, the file will be corrupted.
func (w *FlatFileWriter) InitEx(dest io.WriterAt,recs int, length int64) {
	w.w = dest
	w.nextRec = recs
	w.length  = length
}

/*
Scans a flat-file until the end, and then use its length and record count to initialize the FFWriter.

It will always return an error, usually EBadRecord or io.EOF, depending on the consistency of the file.
If the file was corrupted, it will read until the last usable record.
After that, it will append new Entries.
*/
func (w *FlatFileWriter) InitAppend(dest ReaderWriter) (err error){
	w.w = dest
	w.nextRec, w.length, err = ScanFlatFile(dest)
	return
}
func (w *FlatFileWriter) Append(buf []byte) (recordID int,err error) {
	rl := len(buf)+4
	pos := w.length
	recordID = w.nextRec
	if rl>MaxRawSize { return 0,ERecordTooLong }
	bE.PutUint32(w.buf[:],uint32(rl))
	_,err = w.w.WriteAt(w.buf[:],pos)
	if err!=nil { return }
	_,err = w.w.WriteAt(buf,pos+4)
	if err!=nil { return }
	w.length = pos + int64(rl)
	w.nextRec = recordID+1
	return
}
func (w *FlatFileWriter) ShouldHaveLength() int64 { return w.length }

type pcPair struct{
	recordID int
	offset   int64
}

type PositionCache  struct{
	pairs []pcPair
	steps int
	max   int
}
func (p *PositionCache) Init(max int) {
	if max<256 { max = 1<<10 }
	p.pairs = p.pairs[:0]
	if cap(p.pairs)<max { p.pairs = make([]pcPair,0,max) }
	p.steps = 1
	p.max   = max
}
func (p *PositionCache) Last() (int,int64,bool){
	pairs := p.pairs
	if len(pairs)==0 { return 0,0,false }
	lp := pairs[len(pairs)-1]
	return lp.recordID,lp.offset,true
}
func (p *PositionCache) compact() {
	l := len(p.pairs)
	for i := 0 ; i<l ; i+=2 {
		p.pairs[i/2] = p.pairs[i]
	}
	p.pairs = p.pairs[:l/2]
}
func (p *PositionCache) Append(id int,off int64) {
	if li,_,lok := p.Last(); lok {
		if (li+p.steps)>id { return } /* Don't append */
	}
	p.pairs = append(p.pairs,pcPair{id,off})
	if len(p.pairs) == p.max { p.compact() }
}
func (p *PositionCache) Search(id int) (int,int64) {
	pairs := p.pairs
	i := sort.Search(len(pairs),func(i int) bool {
		return pairs[i].recordID>=id
	})
	if i>=len(pairs) { i = len(pairs)-1 }
	return pairs[i].recordID,pairs[i].offset
}


type FlatFileReader struct{
	r   io.ReaderAt
	pc  PositionCache
	mtx sync.RWMutex
}
func (r *FlatFileReader) Init(src io.ReaderAt) {
	r.r = src
	r.pc.Init(0)
}
func (r *FlatFileReader) InitEx(src io.ReaderAt,maxCache int) {
	r.r = src
	r.pc.Init(maxCache)
}
func (r *FlatFileReader) lookupCache(recordID int) (int,int64,bool) {
	last,loff,lok := r.pc.Last()
	if !lok || last>recordID { return last,loff,true }
	if last==recordID { return last,loff,false }
	fid,foff := r.pc.Search(recordID)
	return fid,foff,false
}
func (r *FlatFileReader) lookup(recordID int) (offset int64,err error){
	r.mtx.RLock()
	id,off,fill := r.lookupCache(recordID)
	r.mtx.RUnlock()
	var buf [4]byte
	if fill { // We have to fill the Cache, so Acquire a writelock
		r.mtx.Lock()
		// Refresh the values.
		id,off,fill = r.lookupCache(recordID)
		if fill {
			defer r.mtx.Unlock()
		} else {
			// If we don't have to fill, Unlock immediately.
			r.mtx.Unlock()
		}
	}
	if id>recordID { id,off = 0,0 }
	for id<recordID {
		_,err = r.r.ReadAt(buf[:],off)
		if err!=nil { return }
		recl := bE.Uint32(buf[:])
		if recl>MaxRawSize || recl<4 { err = EBadRecord ; return }
		off += int64(recl)
		id++
		if fill { r.pc.Append(id,off) }
	}
	offset = off
	return
}
func (r *FlatFileReader) FillCache(count int) {
	r.mtx.Lock() ; defer r.mtx.Unlock()
	id,off,_ := r.pc.Last()
	var buf [4]byte
	for count>0 {
		count--
		_,err := r.r.ReadAt(buf[:],off)
		if err!=nil { return }
		recl := bE.Uint32(buf[:])
		if recl>MaxRawSize || recl<4 { return }
		off += int64(recl)
		id++
		r.pc.Append(id,off)
	}
}
func (r *FlatFileReader) ReadEntry(recordID int) (*[]byte,[]byte,error) {
	offset,err := r.lookup(recordID)
	if err!=nil { return nil,nil,err }
	var buf [4]byte
	_,err = r.r.ReadAt(buf[:],offset)
	if err!=nil { return nil,nil,err }
	recl := bE.Uint32(buf[:])
	if recl>MaxRawSize || recl<4 { return nil,nil,EBadRecord }
	recl-=4
	bobj := buffer.Get(int(recl))
	_,err = r.r.ReadAt((*bobj)[:recl],4+offset)
	if err!=nil {
		buffer.Put(bobj)
		return nil,nil,err
	}
	return bobj,(*bobj)[:recl],nil
}
func (r *FlatFileReader) ReadPosition(recordID int) (recordOffset int64,recordLength int,ioError error) {
	offset,err := r.lookup(recordID)
	if err!=nil { return 0,0,err }
	var buf [4]byte
	_,err = r.r.ReadAt(buf[:],offset)
	if err!=nil { return 0,0,err }
	recl := bE.Uint32(buf[:])
	if recl>MaxRawSize || recl<4 { return 0,0,EBadRecord }
	recl-=4
	return offset,int(recl),nil
}

type FlatFileIterator struct{
	r    io.ReaderAt
	recs int
	off  int64
}
func (f *FlatFileIterator) Init(src io.ReaderAt) {
	f.r    = src
	f.recs = 0
	f.off  = 0
}
func (f *FlatFileIterator) Next() (recordID int,recordOffset int64,recordLength int,ioError error) {
	var buf [4]byte
	off := f.off
	_,err := f.r.ReadAt(buf[:],off)
	if err!=nil { return 0,0,0,err }
	recl := bE.Uint32(buf[:])
	if recl>MaxRawSize || recl<4 { return 0,0,0,EBadRecord }
	f.off = off+int64(recl)
	f.recs++
	return f.recs-1,off+4,int(recl-4),nil
}

