// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package array

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/apache/arrow/go/v9/arrow"
	"github.com/apache/arrow/go/v9/arrow/memory"
	"github.com/goccy/go-json"
)

// String represents an immutable sequence of variable-length UTF-8 strings.
type String struct {
	array
	offsets []int32
	values  string
}

// NewStringData constructs a new String array from data.
func NewStringData(data arrow.ArrayData) *String {
	a := &String{}
	a.refCount = 1
	a.setData(data.(*Data))
	return a
}

// Reset resets the String with a different set of Data.
func (a *String) Reset(data arrow.ArrayData) {
	a.setData(data.(*Data))
}

// Value returns the slice at index i. This value should not be mutated.
func (a *String) Value(i int) string {
	i = i + a.array.data.offset
	return a.values[a.offsets[i]:a.offsets[i+1]]
}

// ValueOffset returns the offset of the value at index i.
func (a *String) ValueOffset(i int) int {
	if i < 0 || i > a.array.data.length {
		panic("arrow/array: index out of range")
	}
	return int(a.offsets[i+a.array.data.offset])
}

func (a *String) ValueOffsets() []int32 {
	beg := a.array.data.offset
	end := beg + a.array.data.length + 1
	return a.offsets[beg:end]
}

func (a *String) ValueBytes() (ret []byte) {
	beg := a.array.data.offset
	end := beg + a.array.data.length
	data := a.values[a.offsets[beg]:a.offsets[end]]

	s := (*reflect.SliceHeader)(unsafe.Pointer(&ret))
	s.Data = (*reflect.StringHeader)(unsafe.Pointer(&data)).Data
	s.Len = len(data)
	s.Cap = len(data)
	return
}

func (a *String) String() string {
	o := new(strings.Builder)
	o.WriteString("[")
	for i := 0; i < a.Len(); i++ {
		if i > 0 {
			o.WriteString(" ")
		}
		switch {
		case a.IsNull(i):
			o.WriteString("(null)")
		default:
			fmt.Fprintf(o, "%q", a.Value(i))
		}
	}
	o.WriteString("]")
	return o.String()
}

func (a *String) setData(data *Data) {
	if len(data.buffers) != 3 {
		panic("arrow/array: len(data.buffers) != 3")
	}

	a.array.setData(data)

	if vdata := data.buffers[2]; vdata != nil {
		b := vdata.Bytes()
		a.values = *(*string)(unsafe.Pointer(&b))
	}

	if offsets := data.buffers[1]; offsets != nil {
		a.offsets = arrow.Int32Traits.CastFromBytes(offsets.Bytes())
	}

	if a.array.data.length < 1 {
		return
	}

	expNumOffsets := a.array.data.offset + a.array.data.length + 1
	if len(a.offsets) < expNumOffsets {
		panic(fmt.Errorf("arrow/array: string offset buffer must have at least %d values", expNumOffsets))
	}

	if int(a.offsets[expNumOffsets-1]) > len(a.values) {
		panic("arrow/array: string offsets out of bounds of data buffer")
	}
}

func (a *String) getOneForMarshal(i int) interface{} {
	if a.IsValid(i) {
		return a.Value(i)
	}
	return nil
}

func (a *String) MarshalJSON() ([]byte, error) {
	vals := make([]interface{}, a.Len())
	for i := 0; i < a.Len(); i++ {
		if a.IsValid(i) {
			vals[i] = a.Value(i)
		} else {
			vals[i] = nil
		}
	}
	return json.Marshal(vals)
}

func arrayEqualString(left, right *String) bool {
	for i := 0; i < left.Len(); i++ {
		if left.IsNull(i) {
			continue
		}
		if left.Value(i) != right.Value(i) {
			return false
		}
	}
	return true
}

// String represents an immutable sequence of variable-length UTF-8 strings.
type LargeString struct {
	array
	offsets []int64
	values  string
}

// NewStringData constructs a new String array from data.
func NewLargeStringData(data arrow.ArrayData) *LargeString {
	a := &LargeString{}
	a.refCount = 1
	a.setData(data.(*Data))
	return a
}

// Reset resets the String with a different set of Data.
func (a *LargeString) Reset(data arrow.ArrayData) {
	a.setData(data.(*Data))
}

// Value returns the slice at index i. This value should not be mutated.
func (a *LargeString) Value(i int) string {
	i = i + a.array.data.offset
	return a.values[a.offsets[i]:a.offsets[i+1]]
}

// ValueOffset returns the offset of the value at index i.
func (a *LargeString) ValueOffset(i int) int64 {
	if i < 0 || i > a.array.data.length {
		panic("arrow/array: index out of range")
	}
	return a.offsets[i+a.array.data.offset]
}

func (a *LargeString) ValueOffsets() []int64 {
	beg := a.array.data.offset
	end := beg + a.array.data.length + 1
	return a.offsets[beg:end]
}

func (a *LargeString) ValueBytes() (ret []byte) {
	beg := a.array.data.offset
	end := beg + a.array.data.length
	data := a.values[a.offsets[beg]:a.offsets[end]]

	s := (*reflect.SliceHeader)(unsafe.Pointer(&ret))
	s.Data = (*reflect.StringHeader)(unsafe.Pointer(&data)).Data
	s.Len = len(data)
	s.Cap = len(data)
	return
}

func (a *LargeString) String() string {
	o := new(strings.Builder)
	o.WriteString("[")
	for i := 0; i < a.Len(); i++ {
		if i > 0 {
			o.WriteString(" ")
		}
		switch {
		case a.IsNull(i):
			o.WriteString("(null)")
		default:
			fmt.Fprintf(o, "%q", a.Value(i))
		}
	}
	o.WriteString("]")
	return o.String()
}

func (a *LargeString) setData(data *Data) {
	if len(data.buffers) != 3 {
		panic("arrow/array: len(data.buffers) != 3")
	}

	a.array.setData(data)

	if vdata := data.buffers[2]; vdata != nil {
		b := vdata.Bytes()
		a.values = *(*string)(unsafe.Pointer(&b))
	}

	if offsets := data.buffers[1]; offsets != nil {
		a.offsets = arrow.Int64Traits.CastFromBytes(offsets.Bytes())
	}

	if a.array.data.length < 1 {
		return
	}

	expNumOffsets := a.array.data.offset + a.array.data.length + 1
	if len(a.offsets) < expNumOffsets {
		panic(fmt.Errorf("arrow/array: string offset buffer must have at least %d values", expNumOffsets))
	}

	if int(a.offsets[expNumOffsets-1]) > len(a.values) {
		panic("arrow/array: string offsets out of bounds of data buffer")
	}
}

func (a *LargeString) getOneForMarshal(i int) interface{} {
	if a.IsValid(i) {
		return a.Value(i)
	}
	return nil
}

func (a *LargeString) MarshalJSON() ([]byte, error) {
	vals := make([]interface{}, a.Len())
	for i := 0; i < a.Len(); i++ {
		if a.IsValid(i) {
			vals[i] = a.Value(i)
		} else {
			vals[i] = nil
		}
	}
	return json.Marshal(vals)
}

func arrayEqualLargeString(left, right *LargeString) bool {
	for i := 0; i < left.Len(); i++ {
		if left.IsNull(i) {
			continue
		}
		if left.Value(i) != right.Value(i) {
			return false
		}
	}
	return true
}

// A StringBuilder is used to build a String array using the Append methods.
type StringBuilder struct {
	*BinaryBuilder
}

// NewStringBuilder creates a new StringBuilder.
func NewStringBuilder(mem memory.Allocator) *StringBuilder {
	b := &StringBuilder{
		BinaryBuilder: NewBinaryBuilder(mem, arrow.BinaryTypes.String),
	}
	return b
}

// Append appends a string to the builder.
func (b *StringBuilder) Append(v string) {
	b.BinaryBuilder.Append([]byte(v))
}

// AppendValues will append the values in the v slice. The valid slice determines which values
// in v are valid (not null). The valid slice must either be empty or be equal in length to v. If empty,
// all values in v are appended and considered valid.
func (b *StringBuilder) AppendValues(v []string, valid []bool) {
	b.BinaryBuilder.AppendStringValues(v, valid)
}

// Value returns the string at index i.
func (b *StringBuilder) Value(i int) string {
	return string(b.BinaryBuilder.Value(i))
}

// NewArray creates a String array from the memory buffers used by the builder and resets the StringBuilder
// so it can be used to build a new array.
func (b *StringBuilder) NewArray() arrow.Array {
	return b.NewStringArray()
}

// NewStringArray creates a String array from the memory buffers used by the builder and resets the StringBuilder
// so it can be used to build a new array.
func (b *StringBuilder) NewStringArray() (a *String) {
	data := b.newData()
	a = NewStringData(data)
	data.Release()
	return
}

func (b *StringBuilder) unmarshalOne(dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}

	switch v := t.(type) {
	case nil:
		b.AppendNull()
	case string:
		b.Append(v)
	default:
		return &json.UnmarshalTypeError{
			Value:  fmt.Sprint(v),
			Type:   reflect.TypeOf(string("")),
			Offset: dec.InputOffset(),
		}
	}
	return nil
}

func (b *StringBuilder) unmarshal(dec *json.Decoder) error {
	for dec.More() {
		if err := b.unmarshalOne(dec); err != nil {
			return err
		}
	}
	return nil
}

func (b *StringBuilder) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil {
		return err
	}

	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("string builder must unpack from json array, found %s", delim)
	}

	return b.unmarshal(dec)
}

// A LargeStringBuilder is used to build a LargeString array using the Append methods.
// LargeString is for when you need the offset buffer to be 64-bit integers
// instead of 32-bit integers.
type LargeStringBuilder struct {
	*BinaryBuilder
}

// NewStringBuilder creates a new StringBuilder.
func NewLargeStringBuilder(mem memory.Allocator) *LargeStringBuilder {
	b := &LargeStringBuilder{
		BinaryBuilder: NewBinaryBuilder(mem, arrow.BinaryTypes.LargeString),
	}
	return b
}

// Append appends a string to the builder.
func (b *LargeStringBuilder) Append(v string) {
	b.BinaryBuilder.Append([]byte(v))
}

// AppendValues will append the values in the v slice. The valid slice determines which values
// in v are valid (not null). The valid slice must either be empty or be equal in length to v. If empty,
// all values in v are appended and considered valid.
func (b *LargeStringBuilder) AppendValues(v []string, valid []bool) {
	b.BinaryBuilder.AppendStringValues(v, valid)
}

// Value returns the string at index i.
func (b *LargeStringBuilder) Value(i int) string {
	return string(b.BinaryBuilder.Value(i))
}

// NewArray creates a String array from the memory buffers used by the builder and resets the StringBuilder
// so it can be used to build a new array.
func (b *LargeStringBuilder) NewArray() arrow.Array {
	return b.NewLargeStringArray()
}

// NewStringArray creates a String array from the memory buffers used by the builder and resets the StringBuilder
// so it can be used to build a new array.
func (b *LargeStringBuilder) NewLargeStringArray() (a *LargeString) {
	data := b.newData()
	a = NewLargeStringData(data)
	data.Release()
	return
}

func (b *LargeStringBuilder) unmarshalOne(dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}

	switch v := t.(type) {
	case nil:
		b.AppendNull()
	case string:
		b.Append(v)
	default:
		return &json.UnmarshalTypeError{
			Value:  fmt.Sprint(v),
			Type:   reflect.TypeOf(string("")),
			Offset: dec.InputOffset(),
		}
	}
	return nil
}

func (b *LargeStringBuilder) unmarshal(dec *json.Decoder) error {
	for dec.More() {
		if err := b.unmarshalOne(dec); err != nil {
			return err
		}
	}
	return nil
}

func (b *LargeStringBuilder) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil {
		return err
	}

	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("string builder must unpack from json array, found %s", delim)
	}

	return b.unmarshal(dec)
}

var (
	_ arrow.Array = (*String)(nil)
	_ arrow.Array = (*LargeString)(nil)
	_ Builder     = (*StringBuilder)(nil)
	_ Builder     = (*LargeStringBuilder)(nil)
)
