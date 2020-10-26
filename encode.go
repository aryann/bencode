package bencode

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"unicode"
)

// Marshal returns a bencode encoding of v.
func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := marshal(reflect.ValueOf(v), &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func marshal(v reflect.Value, buf *bytes.Buffer) error {
	var err error
	switch v.Kind() {
	case reflect.Interface:
		err = marshal(v.Elem(), buf)
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		marshalInt(int(v.Int()), buf)
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		marshalInt(int(v.Uint()), buf)
	case reflect.String:
		err = marshalString(v.String(), buf)
	case reflect.Array, reflect.Slice:
		err = marshalList(v, buf)
	case reflect.Struct:
		err = marshalStruct(v, buf)
	default:
		return fmt.Errorf("encountered unsupported type: %s", v.Kind().String())
	}
	return err
}

func marshalInt(i int, buf *bytes.Buffer) {
	buf.WriteRune('i')
	buf.WriteString(strconv.Itoa(i))
	buf.WriteRune('e')
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func marshalString(s string, buf *bytes.Buffer) error {
	if !isASCII(s) {
		return fmt.Errorf("strings may not contain non-ascii characters: %s", s)
	}
	buf.WriteString(strconv.Itoa(len(s)))
	buf.WriteRune(':')
	buf.WriteString(s)
	return nil
}

func marshalList(v reflect.Value, buf *bytes.Buffer) error {
	buf.WriteRune('l')
	for i := 0; i < v.Len(); i++ {
		if err := marshal(v.Index(i), buf); err != nil {
			return err
		}
	}
	buf.WriteRune('e')
	return nil
}

// marshalStruct serializes a struct. Each field in the struct must have a
// tag named "key" that specifies the key to use in the output. Per Bencode
// specifications, the keys are ordered in the serialized output.
func marshalStruct(v reflect.Value, buf *bytes.Buffer) error {
	keys := make([]string, v.NumField())
	keyToIndex := make(map[string]int, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		key := v.Type().Field(i).Tag.Get("key")
		if key == "" {
			return fmt.Errorf("found struct field with no 'key' tag")
		}
		keys[i] = key
		keyToIndex[key] = i
	}
	sort.Strings(keys)

	buf.WriteRune('d')
	for _, key := range keys {
		if err := marshalString(key, buf); err != nil {
			return err
		}
		marshal(v.Field(keyToIndex[key]), buf)
	}
	buf.WriteRune('e')
	return nil
}
