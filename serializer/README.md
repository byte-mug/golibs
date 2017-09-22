# Serializer

A reflection-based deterministic serialization and deserialization framework build around PreciseIO.

### Define your model.

```go
type Foo struct{
	Naming int
	Content string
}

// Serializer supports automatic
var ser_Foo = serializer.With(new(Foo)).
	Field("Naming").
	Field("Content")

type Bar struct{
	Slice []int
	Map   map[int]int
	Array [2]int
}

/*
 * Serializer supports Slices, Maps and Arrays as serialization Format.
 */
var ser_Bar = serializer.With(new(Bar)).
	Field("Slice").
	Field("Map").
	Field("Array")

type Car struct{
	SV [][]*Bar
}

var ser_Car = serializer.With(new(Car)).
	// Easy support for cascaded containers
	FieldContainerWithDepth("SV",2,ser_Bar)

/*
 * ser_Swtc distinguishes between *Foo, *Bar and *Car
 */
var ser_Swtc = serializer.Switch('0').
	AddTypeWith('F',new(Foo),ser_Foo).
	AddTypeWith('B',new(Bar),ser_Bar).
	AddTypeWith('C',new(Car),ser_Car)

```

### Serialize / Deserialize

```go
func serialize(bwr *bufio.Writer, value interface{}) error {
	w := new(preciseio.PreciseWriter)
	w.Initialize()
	w.W = bwr
	
	return serializer.Serialize(ser_Swtc,w,value)
}
func deserialize(br *bufio.Reader) {
	r := preciseio.PreciseReader{br}
	
	fmt.Println(serializer.Deserialize(ser_Swtc,r))
}
```

