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
		return fmt.Errorf("trailing data at offset %d is not parsable", i)
	}
	return nil
}

func unmarshalNext(offset int, data string, value reflect.Value) (int, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("no data to read at offset %d", offset)
	}

	switch data[offset] {
	case integer:
		return unmarshalInt(offset, data, value)
	case list:
		return unmarshalList(offset, data, value)
	default:
		return 0, fmt.Errorf("expected start of integer, string, list, or dictionary at offset %d", offset)
	}
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func unmarshalInt(offset int, data string, value reflect.Value) (int, error) {
	intStart := offset + 1
	intLimit := intStart + 1
	for intLimit < len(data) && isDigit(data[intLimit]) {
		intLimit++
	}

	i, err := strconv.Atoi(data[intStart:intLimit])
	if err != nil {
		return 0, fmt.Errorf("expected integer at offset %d", intStart)
	}

	if intLimit >= len(data) {
		return 0, fmt.Errorf("expected integer terminator at offset %d", intLimit)
	}
	if data[intLimit] != terminator {
		return 0, fmt.Errorf("expected terminator '%s' for integer at offset %d", string(terminator), offset)
	}

	value.Elem().SetInt(int64(i))
	return intLimit + 1, nil
}

func unmarshalList(offset int, data string, value reflect.Value) (int, error) {
	elemType := value.Type().Elem().Elem()

	offset++ // Consume 'l'.
	for offset < len(data) && data[offset] != terminator {
		newValue := reflect.New(elemType)

		newOffset, err := unmarshalNext(offset, data, newValue)
		if err != nil {
			return 0, err
		}
		offset = newOffset
		value.Elem().Set(reflect.Append(reflect.Indirect(value), reflect.Indirect(newValue)))
	}

	if offset >= len(data) || data[offset] != terminator {
		return 0, fmt.Errorf("expected terminator for list at offset %d", offset)
	}
	return offset + 1, nil
}