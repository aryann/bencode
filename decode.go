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

	i, err := unmarshalNext(0, data, value)
	if err != nil {
		return err
	}
	if i < len(data) {
		return fmt.Errorf("trailing data at offset %d cannot be parsed", i)
	}
	return nil
}

func unmarshalNext(offset int, data string, value reflect.Value) (int, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("no data to read at offset %d", offset)
	}

	if isDigit(data[offset]) {
		return unmarshalString(offset, data, value)
	}

	switch data[offset] {
	case integer:
		return unmarshalInt(offset, data, value)
	case list:
		return unmarshalList(offset, data, value)
	case dictionary:
		return unmarshalDict(offset, data, value)
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

func unmarshalString(offset int, data string, value reflect.Value) (int, error) {
	start, limit, err := stringIndices(offset, data)
	if err != nil {
		return 0, err
	}
	value.Elem().SetString(data[start:limit])
	return limit, nil
}

func unmarshalInt(offset int, data string, value reflect.Value) (int, error) {
	intStart := offset + 1
	intLimit := intLimit(intStart+1, data) // First character may be a '-'.

	i, err := strconv.Atoi(data[intStart:intLimit])
	if err != nil {
		return 0, fmt.Errorf("expected integer at offset %d", intStart)
	}

	if intLimit >= len(data) || data[intLimit] != terminator {
		return 0, fmt.Errorf("expected terminator for integer at offset %d", intLimit)
	}

	value.Elem().SetInt(int64(i))
	return intLimit + 1, nil
}

func unmarshalList(offset int, data string, value reflect.Value) (int, error) {
	offset++ // Consume 'l'.
	elemType := value.Elem().Type().Elem()

	for offset < len(data) && data[offset] != terminator {
		newValue := reflect.New(elemType)

		newOffset, err := unmarshalNext(offset, data, newValue)
		if err != nil {
			return 0, err
		}
		offset = newOffset
		value.Elem().Set(reflect.Append(value.Elem(), reflect.Indirect(newValue)))
	}

	if offset >= len(data) || data[offset] != terminator {
		return 0, fmt.Errorf("expected terminator for list at offset %d", offset)
	}
	return offset + 1, nil
}

func unmarshalDict(offset int, data string, value reflect.Value) (int, error) {
	structValue := value.Elem()
	structType := structValue.Type()
	structValues := make(map[string]reflect.Value)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		key, ok := field.Tag.Lookup("key")
		if !ok {
			continue
		}
		structValues[key] = structValue.Field(i)
	}

	offset++ // Consume 'd'.
	for offset < len(data) && data[offset] != terminator {
		if !isDigit(data[offset]) {
			return 0, fmt.Errorf("dictionary key at offset %d is not a string", offset)
		}
		start, limit, err := stringIndices(offset, data)
		if err != nil {
			return 0, err
		}
		key := data[start:limit]
		value, ok := structValues[key]
		if !ok {
			// TODO: This is too restrictive. We should just ignore
			// unrecognized keys much the same way the json package does.
			return 0, fmt.Errorf("dictionary contains key '%s' at offset %d which does not exist in the given struct", key, offset)
		}

		valueOffset := limit
		newOffset, err := unmarshalNext(valueOffset, data, value)
		if err != nil {
			return 0, err
		}
		offset = newOffset
	}

	if offset >= len(data) || data[offset] != terminator {
		return 0, fmt.Errorf("expected terminator for dictionary at offset %d", offset)
	}
	return offset + 1, nil
}
