# golibs
This Go Library contains a larger set of useful Go packages for different purposes.

[![GoDoc](https://byte-mug.github.io/pkg/godoc-status.svg)](https://byte-mug.github.io/pkg/github.com/byte-mug/golibs)

## Packages

### VersionVector

A Version Vector for concurrency control. https://en.wikipedia.org/wiki/Version_vector

### PreciseIO

Extended IO Routines to construct Serializers/Deserializers to directly operate on
`*bufio.Reader` and `*bufio.Writer`.

### Serializer

A reflection-based deterministic serialization and deserialization framework build around PreciseIO.

Further informations [here](serializer/). [GoDoc](https://byte-mug.github.io/pkg/github.com/byte-mug/golibs/serializer/)

### QuickDump

Another reflection-based (less) deterministic serialization and deserialization framework build around PreciseIO.

Easier to use than Serializer. [GoDoc](https://byte-mug.github.io/pkg/github.com/byte-mug/golibs/quickdump/)

### PStruct

A structure reading and writing library similar to "encoding/binary"

[GoDoc](https://byte-mug.github.io/pkg/github.com/byte-mug/golibs/pstruct/)

### Base128

An encoding similar to base64 but it stores 7 bit payload per byte. It uses bytes in the range 128-255.

[GoDoc](https://byte-mug.github.io/pkg/github.com/byte-mug/golibs/base128/)

### Chordhash

Algroithms related to consistent hashing and the Chord DHT algorithm/protocol.

Further informations [here](chordhash/). [GoDoc](https://byte-mug.github.io/pkg/github.com/byte-mug/golibs/chordhash/)

### Skiplist

A Skiplist derived from this [neat project](https://github.com/kkdai/skiplist), but using non-integer keys.

Further informations [here](skiplist/). [GoDoc](https://byte-mug.github.io/pkg/github.com/byte-mug/golibs/skiplist/)

### Concurrent Collections:

* [Concurrent Skiplist](concurrent/skiplist/).
* Yet another [Concurrent Skiplist](concurrent/sortlist/).

