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

// This package offers space-management of os.File-like entities.
//
// Each allocation is broken down to a set of simple, atomic steps, so that
// the file does not get corrupted, if the process or system crashes during an
// allocation or deallocation.
package filealloc

import "github.com/byte-mug/golibs/pstruct"
import "github.com/byte-mug/golibs/buffer"
import "encoding/binary"
import "fmt"

var EInvalidOffset = fmt.Errorf("Invalid Offset")
var EDoubleFree    = fmt.Errorf("Warning: Double Free")
var EInternalError = fmt.Errorf("Invalid Offset")

var bE = binary.BigEndian

const (
	ranks = 30
	minEven = 0x200 // = 512
	minOdd  = 0x300 // = 768 = 512*1.5
)

func roundUp(i int64) int64{
	return (i+0x1ff) & ^int64(0x1ff)
}

func minRank(l int) uint {
	i := uint(0)
	for ; i<15 ; i++ {
		if l<=(minEven<<i) { return i<<1 }
		if l<=(minOdd<<i) { return (i<<1)|1 }
		
	}
	return ranks // Invalid rank
}
func rank2Size(r uint) int64 {
	if (r&1)==0 {
		return minEven<<(r>>1)
	}
	return minOdd<<(r>>1)
}

func maxRank(l int64) uint {
	r := uint(ranks)
	for r>0 {
		r--
		if l>=rank2Size(r) { return r }
	}
	return ranks
}

type page struct {
	Next     int64
	Rank     uint8
	UsedRank uint8
	Status   uint8 // 0 = not linked; 1 = linked
	_        uint8 // Padding to make it 16 byte-align
	_2       uint32
}
var szPage = pstruct.Sizeof(page{})

type stats struct {
	_ [16]byte // Unused space at begin of the file
	
	// Count of Ranks (free-count)
	Npages [ranks]int64
}
var szStats = pstruct.Sizeof(stats{})

type file struct {
	_ [16]byte // Unused space at begin of the file
	
	// Ranks (free-lists)
	Pages [ranks]int64
}
var szFile = pstruct.Sizeof(file{})

type memFile struct {
	f File
	file
	dirty bool
}
func (m *memFile) flush() error {
	if !m.dirty { return nil }
	b := buffer.Get(szFile)
	defer buffer.Put(b)
	pstruct.Write(&(m.file),*b,bE)
	_,e := m.f.WriteAt((*b)[16:szFile],16)
	m.dirty = false
	if e==nil { e = m.f.Sync() }
	return e
}
func (m *memFile) load(f File) error {
	m.f      = f
	b := buffer.Get(szFile)
	defer buffer.Put(b)
	_,e := m.f.ReadAt((*b)[16:szFile],16)
	pstruct.Read(&(m.file),*b,bE)
	return e
}
func (m *memFile) getRank(i uint) (*memPage,error) {
	if i>=ranks { return nil,nil }
	off := m.Pages[i]
	if off<512 { return nil,nil } // off == NULL or off is invalid
	p := new(memPage)
	p.f = m.f
	p.referer = m
	p.refererIdx = i
	err := p.load(m.f,off)
	if err!=nil { return nil,err }
	return p,nil
}

type memPage struct {
	f File
	page
	dirty  bool
	offset int64
	referer    interface{}
	refererIdx uint
}
func (m *memPage) flush() error {
	if !m.dirty { return nil }
	b := buffer.Get(szPage)
	defer buffer.Put(b)
	pstruct.Write(&(m.page),*b,bE)
	_,e := m.f.WriteAt((*b)[:szPage],m.offset)
	m.dirty = false
	if e==nil { e = m.f.Sync() }
	return e
}
func (m *memPage) load(f File,offset int64) error {
	m.f      = f
	m.offset = offset
	b := buffer.Get(szPage)
	defer buffer.Put(b)
	_,e := m.f.ReadAt((*b)[:szPage],m.offset)
	pstruct.Read(&(m.page),*b,bE)
	return e
}
func (m *memPage) unlink() (bool,error) {
	switch v := m.referer.(type) {
	case *memFile:
		if m.refererIdx>=ranks { return false,nil }
		v.Pages[m.refererIdx] = m.Next
		v.dirty = true
		m.referer = nil
		m.Next = 0
		m.Status = 1
		m.dirty = true
		return true,v.flush()
	case *memPage:
		v.Next = m.Next
		v.dirty = true
		m.Next = 0
		m.Status = 1
		m.dirty = true
		return true,v.flush()
	}
	return false,nil
}

type memStats struct {
	f File
	stats
	dirty bool
}
func (m *memStats) flush() error {
	if !m.dirty { return nil }
	b := buffer.Get(szStats)
	defer buffer.Put(b)
	pstruct.Write(&(m.stats),*b,bE)
	_,e := m.f.WriteAt((*b)[:szStats],256)
	m.dirty = false
	if e==nil { e = m.f.Sync() }
	return e
}
func (m *memStats) load(f File) error {
	m.f      = f
	b := buffer.Get(szStats)
	defer buffer.Put(b)
	_,e := m.f.ReadAt((*b)[:szStats],256)
	pstruct.Read(&(m.stats),*b,bE)
	return e
}


// Allocator manages allocation of file blocks within a File.
//
// Allocator methods are not safe for concurrent use by multiple goroutines.
// Callers must provide their own synchronization whan it's used concurrently
// by multiple goroutines.
type Allocator struct{
	f   File
	m   *memFile
	s   *memStats
	eof int64
}
func NewAllocator(f File) (*Allocator,error) {
	return newAllocator(f,false)
}
func newAllocator(f File, repair bool) (*Allocator,error) {
	fi,e := f.Stat()
	if e!=nil { return nil,e }
	
	a := new(Allocator)
	a.f = f
	a.m = new(memFile)
	a.s = new(memStats)
	a.eof = fi.Size()
	
	if a.eof<512 {
		a.eof = 512
		a.m.f = f
		a.m.dirty = true
		a.s.f = f
		a.s.dirty = true
		e := a.m.flush()
		if e!=nil { return nil,e }
		e = a.s.flush()
		if e!=nil { return nil,e }
		return a,nil
	}
	
	e = a.m.load(a.f)
	if e!=nil { return nil,e }
	e = a.s.load(a.f)
	if e!=nil { return nil,e }
	
	if repair {
		panic("Not supported")
		/*for i,off := range a.m.Pages {
			if !a.check(off) { a.m.Pages[i] = 0 } // Drop invalid references.
		}*/
	} else {
		for _,off := range a.m.Pages {
			if !a.check(off) { return nil,fmt.Errorf("corrupted file") }
		}
	}
	
	return a,nil
}
func (a *Allocator) FileSize() int64 { return a.eof }
func (a *Allocator) check(off int64) bool {
	return off<=a.eof && off>=0
}
func (a *Allocator) appendPage(r uint) (*memPage,error) {
	if r==1 { r=2 } // disable rank 1 in favor of rank 2
	var empty [8]byte
	beg := roundUp(a.eof)
	lng := rank2Size(r)
	_,e := a.f.WriteAt(empty[:],lng+beg)
	if e!=nil { return nil,e }
	pg := new(memPage)
	pg.f = a.f
	pg.offset = beg
	pg.Rank = uint8(r)
	pg.UsedRank = uint8(r)
	pg.Status = 1
	pg.dirty = true
	e = pg.flush()
	if e!=nil { return nil,e }
	return pg,nil
}
func (a *Allocator) allocAlgorithm1(i int,noGrow bool) (int64,error) {
	r := minRank(i+16)
	if r>=ranks { return -1,nil } // Chunk too big.
	pg,err := a.m.getRank(r)
	if err!=nil { return -1,err }
	if pg!=nil {
		ok,err := pg.unlink()
		if err!=nil { return -1,err }
		if !ok { return -1,fmt.Errorf("Could not unlink") }
		err = pg.flush()
		if err!=nil { return -1,err }
		a.s.Npages[r]--
		a.s.dirty = true
		a.s.flush()
		return pg.offset+16,nil
	}
	if noGrow { return -1,nil } // No growth
	pg,err = a.appendPage(r)
	if err!=nil { return -1,err }
	
	return pg.offset+16,nil
}
func (a *Allocator) splitOff(pg *memPage) {
	used,rank := pg.UsedRank,pg.Rank
	if rank<2 { return }
	if used==1 { used = 2 }
	if used >= rank { return }
	if (rank&1)==1 {
		for {
			if (rank-2)<used { break }
			if (rank-2)==1 { break }
			rank-=2
			pg2 := &memPage{f:pg.f}
			pg2.Rank = rank
			pg2.offset = pg.offset+rank2Size(uint(rank))
			pg2.dirty = true
			pg2.Next = a.m.Pages[rank-2]
			e := pg2.flush()
			if e!=nil { return }
			
			a.m.Pages[rank] = pg2.offset
			a.m.dirty = true
			e = a.m.flush()
			if e!=nil { return }
			a.s.Npages[rank]++
			
			pg.Rank = rank
		}
	} else {
		// Algorithm for splitting even pages.
		for rank>=2 {
			if (rank-2)<used { break }
			rank-=2
			pg2 := &memPage{f:pg.f}
			pg2.Rank = rank
			pg2.offset = pg.offset+rank2Size(uint(rank))
			pg2.dirty = true
			pg2.Next = a.m.Pages[rank]
			e := pg2.flush()
			if e!=nil { return }
			
			a.m.Pages[rank] = pg2.offset
			a.m.dirty = true
			e = a.m.flush()
			if e!=nil { return }
			a.s.Npages[rank]++
			pg.Rank = rank
		}
		if (rank-1)==used && rank>=4 {
			// size(rank)   = 1/1 = 2/2
			// size(rank-1) = 1.5/2
			// size(rank-2) = 1/2
			// size(rank-4) = 1/4 = 0.5/2
			// size(rank-1) + size(rank-4) = size(rank)
			
			pg2 := &memPage{f:pg.f}
			pg2.Rank = rank-4
			pg2.offset = pg.offset+rank2Size(uint(rank-1))
			pg2.dirty = true
			pg2.Next = a.m.Pages[rank-4]
			e := pg2.flush()
			if e!=nil { return }
			
			a.m.Pages[rank-4] = pg2.offset
			a.m.dirty = true
			e = a.m.flush()
			if e!=nil { return }
			a.s.Npages[rank-4]++
			
			rank--
			pg.Rank = rank
		}
	}
	
}
func (a *Allocator) allocAlgorithm2(i int,noGrow bool) (int64,error) {
	r := minRank(i+16)
	if r>=ranks { return -1,nil } // Chunk too big.
	pg,err := a.m.getRank(r)
	if err!=nil { return -1,err }
	if pg!=nil {
		ok,err := pg.unlink()
		if err!=nil { return -1,err }
		if !ok { return -1,fmt.Errorf("Could not unlink") }
		err = pg.flush()
		if err!=nil { return -1,err }
		a.s.Npages[r]--
		a.s.dirty = true
		a.s.flush()
		return pg.offset+16,nil
	}
	for i := r+1 ; i<ranks ; i++ {
		pg,err := a.m.getRank(i)
		if err!=nil { return -1,err }
		if pg!=nil {
			ok,err := pg.unlink()
			if err!=nil { return -1,err }
			if !ok { return -1,fmt.Errorf("Could not unlink") }
			pg.UsedRank = uint8(r)
			a.splitOff(pg)
			err = pg.flush()
			if err!=nil { return -1,err }
			a.s.Npages[i]--
			a.s.dirty = true
			a.s.flush()
			return pg.offset+16,nil
		}
	}
	fmt.Println("Growth")
	if noGrow { return -1,nil } // No growth
	pg,err = a.appendPage(r)
	if err!=nil { return -1,err }
	
	return pg.offset+16,nil
}

func (a *Allocator) Alloc(size int,noGrow bool) (int64,error) {
	return a.allocAlgorithm2(size,noGrow)
}
func (a *Allocator) Free(off int64) error {
	off-=16
	if (off&0x1ff)!=0 || !a.check(off) { return EInvalidOffset }
	m := new(memPage)
	err := m.load(a.f,off)
	if err!=nil { return err }
	if m.Status==0 { return EDoubleFree }
	if m.Rank>=ranks { return EInternalError }
	
	m.Next = a.m.Pages[m.Rank]
	m.dirty = true
	m.Status = 0
	m.UsedRank = m.Rank
	err = m.flush()
	if err!=nil { return err }
	
	a.m.Pages[m.Rank] = off
	a.m.dirty = true
	a.s.Npages[m.Rank]++
	a.s.dirty = true
	err = a.m.flush()
	a.s.flush()
	return err
}

func (a *Allocator) UsableSize(off int64) (int,error) {
	off-=16
	if (off&0x1ff)!=0 || !a.check(off) { return 0,EInvalidOffset }
	m := new(memPage)
	err := m.load(a.f,off)
	if err!=nil { return 0,err }
	i := int(rank2Size(uint(m.UsedRank)))-16
	return i,nil
}
func (a *Allocator) ApproxFreeSpace() (total int64) {
	for i := uint(0) ; i<ranks ; i++ {
		p := a.s.Npages[i]
		if p<0 { continue }
		total += p*(rank2Size(i)-16)
	}
	return
}
func (a *Allocator) ApproxFreeSpaceFor(minSize int) (total int64) {
	for i := minRank(minSize+16) ; i<ranks ; i++ {
		p := a.s.Npages[i]
		if p<0 { continue }
		total += p*(rank2Size(i)-16)
	}
	return
}
func (a *Allocator) LL_getRanks() [30]int64 {
	return a.s.Npages
}

func LL_getRawSizeForRank(i uint) int {
	return int(rank2Size(i))
}

