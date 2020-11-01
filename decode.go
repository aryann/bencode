package bencode

import (
	"fmt"
	"reflect"
	"strconv"
)

const (
	integer    = 'i'
	list       = 'l'
	dictionary = 'd'
	terminator = 'e'
)

// Unmarshal deserializes a Bencode string.
func Unmarshal(data []byte, v interface{}) error {
	// TODO: Don't modify the interface until we know the full output is valid.

	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return fmt.Errorf("v is not a non-nil pointer: %s", reflect.TypeOf((v)))
	}

	// First run through the input using a no-op valueSetter. This allows us
	// to report an error if the input in malformed without making any partial
	// modifications to the output parameter v.
	validator := decoder{
		data:        data,
		offset:      0,
		valueSetter: noOpValueSetter{},
	}
	err := validator.unmarshalNext(&value)
	if err != nil {
		return err
	}
	if !validator.isDone() {
		return fmt.Errorf("trailing data at offset %d cannot be parsed", validator.offset)
	}

	// The input is valid, so now we do our second pass over the input and
	// fill the output parameter.
	decoder := decoder{
		data:        data,
		offset:      0,
		valueSetter: valueSetter{},
	}
	return decoder.unmarshalNext(&value)
}

// valueSetterInterface abstracts a subset of the reflect.Value modifiers.
type valueSetterInterface interface {
	SetInt(value *reflect.Value, i int64)
	SetString(value *reflect.Value, s string)
	Append(target *reflect.Value, elem reflect.Value)
}

// valueSetter delegates directly to the reflect.Value modifiers.
type valueSetter struct{}

func (valueSetter) SetInt(value *reflect.Value, i int64) {
	value.Elem().SetInt(i)
}
func (valueSetter) SetString(value *reflect.Value, s string) {
	value.Elem().SetString(s)
}
func (valueSetter) Append(target *reflect.Value, elem reflect.Value) {
	target.Elem().Set(reflect.Append(target.Elem(), reflect.Indirect(elem)))
}

// noOpValueSetter is a valueSetterInterface that does nothing. This is useful
// during the validation phase of deserialization.
type noOpValueSetter struct{}

func (noOpValueSetter) SetInt(value *reflect.Value, i int64)             {}
func (noOpValueSetter) SetString(value *reflect.Value, s string)         {}
func (noOpValueSetter) Append(target *reflect.Value, elem reflect.Value) {}

type decoder struct {
	data        []byte
	offset      int
	valueSetter valueSetterInterface
}

func (d *decoder) isDone() bool {
	return len(d.data) <= d.offset
}

func (d *decoder) unmarshalNext(value *reflect.Value) error {
	if len(d.data) == 0 {
		return fmt.Errorf("no data to read at offset %d", d.offset)
	}

	if isDigit(d.data[d.offset]) {
		return d.unmarshalString(value)
	}

	switch d.data[d.offset] {
	case integer:
		return d.unmarshalInt(value)
	case list:
		return d.unmarshalList(value)
	case dictionary:
		return d.unmarshalDict(value)
	}
	return fmt.Errorf("expected start of integer, string, list, or dictionary at offset %d", d.offset)
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func intLimit(offset int, data []byte) int {
	for offset < len(data) && isDigit(data[offset]) {
		offset++
	}
	return offset
}

func stringIndices(offset int, data []byte) (int, int, error) {
	intStart := offset
	intLimit := intLimit(intStart, data)
	length, err := strconv.Atoi(string(data[intStart:intLimit]))
	if err != nil {
		return 0, 0, fmt.Errorf("could not parse length for string at offset %d", offset)
	}
	if intLimit >= len(data) || data[intLimit] != ':' {
		return 0, 0, fmt.Errorf("expected colon between length and value for string at offset %d", offset)
	}
	strStart := intLimit + 1
	strLimit := strStart + length
	if strLimit > len(data) {
		return 0, 0, fmt.Errorf("string at offset %d has length %d, yet there are not that many bytes left", offset, length)
	}
	return strStart, strLimit, nil
}

func (d *decoder) unmarshalString(value *reflect.Value) error {
	start, limit, err := stringIndices(d.offset, d.data)
	if err != nil {
		return err
	}

	if value != nil {
		if value.Elem().Type().Kind() != reflect.String {
			return fmt.Errorf("cannot unmarshal string at offset %d into %s", d.offset, value.Elem().Type())
		}
		d.valueSetter.SetString(value, string(d.data[start:limit]))
	}
	d.offset = limit
	return nil
}

func (d *decoder) unmarshalInt(value *reflect.Value) error {
	intStart := d.offset + 1
	intLimit := intLimit(intStart+1, d.data) // First character may be a '-'.

	i, err := strconv.Atoi(string(d.data[intStart:intLimit]))
	if err != nil {
		return fmt.Errorf("expected integer at offset %d", intStart)
	}

	if intLimit >= len(d.data) || d.data[intLimit] != terminator {
		return fmt.Errorf("expected terminator for integer at offset %d", intLimit)
	}

	if value != nil {
		if value.Elem().Type().Kind() != reflect.Int64 {
			return fmt.Errorf("cannot unmarshal integer at offset %d into %s", d.offset, value.Elem().Type())
		}
		d.valueSetter.SetInt(value, int64(i))
	}
	d.offset = intLimit + 1
	return nil
}

func (d *decoder) unmarshalList(value *reflect.Value) error {
	if value != nil && value.Elem().Type().Kind() != reflect.Slice {
		return fmt.Errorf("cannot unmarshal list at offset %d into %s", d.offset, value.Elem().Type())
	}

	d.offset++ // Consume 'l'.

	for d.offset < len(d.data) && d.data[d.offset] != terminator {
		if value == nil {
			if err := d.unmarshalNext(nil); err != nil {
				return err
			}
			continue
		}

		elem := reflect.New(value.Elem().Type().Elem())
		if err := d.unmarshalNext(&elem); err != nil {
			return err
		}
		d.valueSetter.Append(value, elem)
	}

	if d.offset >= len(d.data) || d.data[d.offset] != terminator {
		return fmt.Errorf("expected terminator for list at offset %d", d.offset)
	}
	d.offset++
	return nil
}

func (d *decoder) unmarshalDict(value *reflect.Value) error {
	if value != nil && value.Elem().Type().Kind() != reflect.Struct {
		return fmt.Errorf("cannot unmarshal dictionary at offset %d into %s", d.offset, value.Elem().Type())
	}

	structValues := make(map[string]reflect.Value)
	if value != nil {
		structType := value.Elem().Type()
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			key, ok := field.Tag.Lookup("bencode")
			if !ok {
				continue
			}
			structValues[key] = value.Elem().Field(i).Addr()
		}
	}

	d.offset++ // Consume 'd'.
	for d.offset < len(d.data) && d.data[d.offset] != terminator {
		if !isDigit(d.data[d.offset]) {
			return fmt.Errorf("dictionary key at offset %d is not a string", d.offset)
		}
		start, limit, err := stringIndices(d.offset, d.data)
		if err != nil {
			return err
		}
		key := string(d.data[start:limit])
		d.offset = limit

		var nextValue *reflect.Value
		value, ok := structValues[key]
		if ok {
			nextValue = &value
		}

		if err := d.unmarshalNext(nextValue); err != nil {
			return err
		}
	}

	if d.offset >= len(d.data) || d.data[d.offset] != terminator {
		return fmt.Errorf("expected terminator for dictionary at offset %d", d.offset)
	}
	d.offset++
	return nil
}
