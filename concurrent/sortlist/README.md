Yet another Concurrent Skiplist in Go
==================
Yet another Concurrent Skiplist implementation. Changes are performed using
atomic CAS operations. Inserts acquire a shared lock and Writes acquire an
exclusive lock. All locking is done internally.


What is a Skip List
-------------------

[Skip List](https://en.wikipedia.org/wiki/Skip_list) is a data structure that allows fast search within an ordered sequence of elements. Fast search is made possible by maintaining a linked hierarchy of subsequences, each skipping over fewer elements. 

![image](https://upload.wikimedia.org/wikipedia/commons/thumb/8/86/Skip_list.svg/500px-Skip_list.svg.png)

(from [wiki](https://en.wikipedia.org/wiki/Skip_list))

Install
---------------
`go get github.com/byte-mug/golibs/concurrent/sortlist`


Usage
---------------

```go
package main


import "fmt"
import "github.com/byte-mug/golibs/concurrent/sortlist"
import "github.com/emirpasic/gods/utils"

func main() {
	
	l := &sortlist.Sortlist{Cmp:utils.IntComparator}
	
	l.Insert(10,nil)
	l.Insert(20,nil)
	l.Insert(30,nil)
	
	fmt.Println(l.Previous(5))
	fmt.Println(l.Previous(15))
	fmt.Println(l.Previous(25))
	fmt.Println(l.Previous(35))
	
	fmt.Println("------------------------------------")
	
	fmt.Println(l.Floor(5))
	fmt.Println(l.Floor(15))
	fmt.Println(l.Floor(25))
	fmt.Println(l.Floor(35))
	
	fmt.Println("------------------------------------")
	
	fmt.Println(l.Next(5))
	fmt.Println(l.Next(15))
	fmt.Println(l.Next(25))
	fmt.Println(l.Next(35))
	
	fmt.Println("------------------------------------")
	
	fmt.Println(l.Ceil(5))
	fmt.Println(l.Ceil(15))
	fmt.Println(l.Ceil(25))
	fmt.Println(l.Ceil(35))
	
	fmt.Println("------------------------------------")
	
	fmt.Println(l.Lookup(5))
	fmt.Println(l.Lookup(10))
	fmt.Println(l.Lookup(20))
	fmt.Println(l.Lookup(30))
	
	fmt.Println("------------------------------------")
	
	fmt.Println()
}
```

### Inspired By:

- [ConcurrentSkipListMap](https://docs.oracle.com/javase/8/docs/api/java/util/concurrent/ConcurrentSkipListMap.html)


License
---------------

This package is licensed under MIT license.


