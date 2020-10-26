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
func Unmarshal(data string, v interface{}) error {
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
		ValueSetter: noOpValueSetter{},
	}
	i, err := validator.unmarshalNext(0, value)
	if err != nil {
		return err
	}
	if i < len(data) {
		return fmt.Errorf("trailing data at offset %d cannot be parsed", i)
	}

	// The input is valid, so now we do our second pass over the input and
	// fill the output parameter.
	decoder := decoder{
		data:        data,
		ValueSetter: valueSetter{},
	}
	_, err = decoder.unmarshalNext(0, value)
	return err
}

// valueSetterInterface abstracts a subset of the reflect.Value modifiers.
type valueSetterInterface interface {
	SetInt(value reflect.Value, i int64)
	SetString(value reflect.Value, s string)
	Append(target reflect.Value, elem reflect.Value)
}

// valueSetter delegates directly to the reflect.Value modifiers.
type valueSetter struct{}

func (valueSetter) SetInt(value reflect.Value, i int64) {
	value.SetInt(i)
}
func (valueSetter) SetString(value reflect.Value, s string) {
	value.SetString(s)
}
func (valueSetter) Append(target reflect.Value, elem reflect.Value) {
	target.Set(reflect.Append(target, reflect.Indirect(elem)))
}

// noOpValueSetter is a valueSetterInterface that does nothing. This is useful
// during the validation phase of deserialization.
type noOpValueSetter struct{}

func (noOpValueSetter) SetInt(value reflect.Value, i int64)             {}
func (noOpValueSetter) SetString(value reflect.Value, s string)         {}
func (noOpValueSetter) Append(target reflect.Value, elem reflect.Value) {}

type decoder struct {
	data        string
	ValueSetter valueSetterInterface
}

func (d *decoder) unmarshalNext(offset int, value reflect.Value) (int, error) {
	value = value.Elem()
	if len(d.data) == 0 {
		return 0, fmt.Errorf("no data to read at offset %d", offset)
	}

	if isDigit(d.data[offset]) {
		return d.unmarshalString(offset, value)
	}

	switch d.data[offset] {
	case integer:
		return d.unmarshalInt(offset, value)
	case list:
		return d.unmarshalList(offset, value)
	case dictionary:
		return d.unmarshalDict(offset, value)
	default:
		return 0, fmt.Errorf("expected start of integer, string, list, or dictionary at offset %d", offset)
	}
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func intLimit(offset int, data string) int {
	for offset < len(data) && isDigit(data[offset]) {
		offset++
	}
	return offset
}

func stringIndices(offset int, data string) (int, int, error) {
	intStart := offset
	intLimit := intLimit(intStart, data)
	length, err := strconv.Atoi(data[intStart:intLimit])
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

func (d *decoder) unmarshalString(offset int, value reflect.Value) (int, error) {
	start, limit, err := stringIndices(offset, d.data)
	if err != nil {
		return 0, err
	}
	d.ValueSetter.SetString(value, d.data[start:limit])
	return limit, nil
}

func (d *decoder) unmarshalInt(offset int, value reflect.Value) (int, error) {
	intStart := offset + 1
	intLimit := intLimit(intStart+1, d.data) // First character may be a '-'.

	i, err := strconv.Atoi(d.data[intStart:intLimit])
	if err != nil {
		return 0, fmt.Errorf("expected integer at offset %d", intStart)
	}

	if intLimit >= len(d.data) || d.data[intLimit] != terminator {
		return 0, fmt.Errorf("expected terminator for integer at offset %d", intLimit)
	}

	d.ValueSetter.SetInt(value, int64(i))
	return intLimit + 1, nil
}

func (d *decoder) unmarshalList(offset int, value reflect.Value) (int, error) {
	offset++ // Consume 'l'.
	elemType := value.Type().Elem()

	for offset < len(d.data) && d.data[offset] != terminator {
		elem := reflect.New(elemType)
		newOffset, err := d.unmarshalNext(offset, elem)
		if err != nil {
			return 0, err
		}
		offset = newOffset
		d.ValueSetter.Append(value, elem)
	}

	if offset >= len(d.data) || d.data[offset] != terminator {
		return 0, fmt.Errorf("expected terminator for list at offset %d", offset)
	}
	return offset + 1, nil
}

func (d *decoder) unmarshalDict(offset int, value reflect.Value) (int, error) {
	structType := value.Type()
	structValues := make(map[string]reflect.Value)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		key, ok := field.Tag.Lookup("key")
		if !ok {
			continue
		}
		structValues[key] = value.Field(i).Addr()
	}

	offset++ // Consume 'd'.
	for offset < len(d.data) && d.data[offset] != terminator {
		if !isDigit(d.data[offset]) {
			return 0, fmt.Errorf("dictionary key at offset %d is not a string", offset)
		}
		start, limit, err := stringIndices(offset, d.data)
		if err != nil {
			return 0, err
		}
		key := d.data[start:limit]
		value, ok := structValues[key]
		if !ok {
			// TODO: This is too restrictive. We should just ignore
			// unrecognized keys much the same way the json package does.
			return 0, fmt.Errorf("dictionary contains key '%s' at offset %d which does not exist in the given struct", key, offset)
		}

		valueOffset := limit
		newOffset, err := d.unmarshalNext(valueOffset, value)
		if err != nil {
			return 0, err
		}
		offset = newOffset
	}

	if offset >= len(d.data) || d.data[offset] != terminator {
		return 0, fmt.Errorf("expected terminator for dictionary at offset %d", offset)
	}
	return offset + 1, nil
}
