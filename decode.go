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
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return fmt.Errorf("v is not a non-nil pointer: %s", reflect.TypeOf((v)))
	}

	return unmarshal(0, data, &value)
}

func unmarshal(offset int, data string, ptr *reflect.Value) error {
	if len(data) == 0 {
		return nil
	}

	switch data[0] {
	case integer:
		return unmarshalInt(offset, data, ptr)
	default:
		return fmt.Errorf("invalid syntax")
	}
}

func unmarshalInt(offset int, data string, ptr *reflect.Value) error {
	value := ptr.Elem()
	if data[len(data)-1] != terminator {
		return fmt.Errorf("found unterminated integer at offset %d", offset)
	}
	data = data[1 : len(data)-1]
	i, err := strconv.Atoi(data)
	if err != nil {
		return fmt.Errorf("could not parse integer at offset %d: %s", offset, data)
	}
	value.SetInt(int64(i))
	return nil
}
