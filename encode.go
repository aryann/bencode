package bencode

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Marshal returns a bencode encoding of v.
func Marshal(v interface{}) (string, error) {
	var buf strings.Builder
	if err := marshal(reflect.ValueOf(v), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func marshal(v reflect.Value, w *strings.Builder) error {
	var err error
	switch v.Kind() {
	case reflect.Interface:
		err = marshal(v.Elem(), w)
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		marshalInt(int(v.Int()), w)
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		marshalInt(int(v.Uint()), w)
	case reflect.String:
		err = marshalString(v.String(), w)
	case reflect.Array, reflect.Slice:
		err = marshalList(v, w)
	case reflect.Struct:
		err = marshalStruct(v, w)
	default:
		return fmt.Errorf("encountered unsupported type: %s", v.Kind().String())
	}
	return err
}

func marshalInt(i int, w *strings.Builder) {
	w.WriteRune('i')
	w.WriteString(strconv.Itoa(i))
	w.WriteRune('e')
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func marshalString(s string, w *strings.Builder) error {
	if !isASCII(s) {
		return fmt.Errorf("strings may not contain non-ascii characters: %s", s)
	}
	w.WriteString(strconv.Itoa(len(s)))
	w.WriteRune(':')
	w.WriteString(s)
	return nil
}

func marshalList(v reflect.Value, w *strings.Builder) error {
	w.WriteRune('l')
	for i := 0; i < v.Len(); i++ {
		if err := marshal(v.Index(i), w); err != nil {
			return err
		}
	}
	w.WriteRune('e')
	return nil
}

// marshalStruct serializes a struct. Each field in the struct must have a
// tag named "key" that specifies the key to use in the output. Per Bencode
// specifications, the keys are ordered in the serialized output.
func marshalStruct(v reflect.Value, w *strings.Builder) error {
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

	w.WriteRune('d')
	for _, key := range keys {
		if err := marshalString(key, w); err != nil {
			return err
		}
		marshal(v.Field(keyToIndex[key]), w)
	}
	w.WriteRune('e')
	return nil
}
