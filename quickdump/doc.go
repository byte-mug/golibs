/*
MIT License

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


/*
QuickDump, a PreciseIO based simple Serialization system, that is much easier to use
than serializer.

QuickDump is capable to serialize structures without any need to previously create serializers for it.

As a limitation: QuickDump is strictly typed - "interface{}" does not work at all. Also
QuickDump is brittle. Wrong types will cause QuickDump to panic.


Nullable

By default every pointer is a NULLable (implicitely). You can cause every NULLable value to be
treated as non-nullable by setting the quickdump:"strip" tag. You can also explicitely mark
a value as NULLable, if it's type supports it, by setting the quickdump:"nullable" tag.

Here are the examples of nullable values:

	Nullable1 *DesiredType                         `quickdump:"nullable"` // NULLable by default, but we made it explicitely.
	Nullable2 struct{ Ok bool; Data DesiredType }  `quickdump:"nullable"` // If Ok is false, value is NULL


Variants

Variants are a sequence of Fields, where ony one of them will be serialized. They are used to implement
some kind of Dynamic Typing/Polymorphism.

	type Variant struct{
		Alpha *Alpha  `quickdump:"tag,strip"`
		Beta  *Beta   `quickdump:"more,strip"`
		Gamma *Gamma  `quickdump:"more,strip"`
	}
*/
package quickdump


