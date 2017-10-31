# golibs
This Go Library contains a larger set of useful Go packages for different purposes.

[![GoDoc](https://godoc.org/github.com/byte-mug/golibs?status.svg)](https://godoc.org/github.com/byte-mug/golibs)

## Packages

### VersionVector

A Version Vector for concurrency control. https://en.wikipedia.org/wiki/Version_vector

### PreciseIO

Extended IO Routines to construct Serializers/Deserializers to directly operate on
`*bufio.Reader` and `*bufio.Writer`.

### Serializer

A reflection-based deterministic serialization and deserialization framework build around PreciseIO.

Further informations [here](serializer/).

### QuickDump

Another reflection-based (less) deterministic serialization and deserialization framework build around PreciseIO.

Easier to use than Serializer.

### PStruct

A structure reading and writing library similar to "encoding/binary"

### Base128

An encoding similar to base64 but it stores 7 bit payload per byte. It uses bytes in the range 128-255.
