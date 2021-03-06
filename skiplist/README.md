Skip List implement in Go
==================
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/byte-mug/golibs/master/skiplist/LICENSE)  [![GoDoc](https://godoc.org/github.com/byte-mug/golibs/skiplist?status.svg)](https://godoc.org/github.com/byte-mug/golibs/skiplist)


What is Skip List
---------------

[Skip List](https://en.wikipedia.org/wiki/Skip_list) is a data structure that allows fast search within an ordered sequence of elements. Fast search is made possible by maintaining a linked hierarchy of subsequences, each skipping over fewer elements. 

![image](https://upload.wikimedia.org/wikipedia/commons/thumb/8/86/Skip_list.svg/500px-Skip_list.svg.png)

(from [wiki](https://en.wikipedia.org/wiki/Skip_list))

Install
---------------
`go get github.com/byte-mug/golibs/skiplist`


Usage
---------------

```go

    package main
    
    import (
	"fmt"
	
	. "github.com/byte-mug/golibs/skiplist"
	"github.com/emirpasic/gods/utils"
    )

    func main() {
        //New a skiplist
        sl := skiplist.NewSkipList(utils.IntComparator)

        //Insert search key 50, value "5", value could be anything.
        sl.Insert(50, "5")
        sl.Insert(40, "4")
        sl.Insert(70, "7")
        sl.Insert(100, "10")

        //Search key, which time complexity O(log n)
        ret, err := sl.Search(50)
        if err == nil {
            fmt.Println("key 50: val->", ret)
        } else {
            fmt.Println("Not found, ", err)
        }

        //Delete by search key
        err = sl.Delete(70)
        if err != nil {
            fmt.Println("Delete not found")
        }

        //Display all skip list content.
        sl.DisplayAll()

    //head->[key:0][val:header]->[key:40][val:4]->[key:50][val:5]->[key:100][val:10]->nil
    //---------------------------------------------------------
    //[node:0], val:header, level:4  fw[3]:40 fw[2]:40 fw[1]:40 fw[0]:40
    //[node:40], val:4, level:4  fw[3]:nil fw[2]:nil fw[1]:50 fw[0]:50
    //[node:50], val:5, level:2  fw[1]:100 fw[0]:100
    //[node:100], val:10, level:2
    }    
```

### Inspired By:

- [Skip List bookbook](http://drum.lib.umd.edu/bitstream/handle/1903/544/CS-TR-2286.1.pdf)
- [SkipList Wiki](https://en.wikipedia.org/wiki/Skip_list)
- [大数据日知录：架构与算法](http://product.dangdang.com/23561651.html)
- [https://github.com/surge/skiplist](https://github.com/surge/skiplist)
- [https://github.com/huandu/skiplist](https://github.com/huandu/skiplist)
- [https://github.com/gansidui/skiplist](https://github.com/gansidui/skiplist)

Benchmark
---------------

Here is the benchmark list:

**Note**: The slice search op base on worst case.



```
//Slice Operations
BenchmarkSliceInsert-4   	100000000	        34.4  ns/op
BenchmarkSliceSearch-4   	   20000	     80149.0  ns/op

//Sliplist Operations
BenchmarkSkiplistInsert-4	  100000	     20881.0  ns/op
BenchmarkSkiplistSearch-4	10000000	       137.0  ns/op
```

*It is very obviousily the skiplist search is much faster than slice iterator.*


The map operation for reference.

```
//Map Operations
BenchmarkMapInsert-4     	 5000000	       283.0  ns/op
BenchmarkMapSearch-4     	30000000	        42.4  ns/op
```

Project52
---------------

It is one of my [project 52](https://github.com/kkdai/project52).


License
---------------

This package is licensed under MIT license. See LICENSE for details.


