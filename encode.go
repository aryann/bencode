package bencode

import (
	"fmt"
	"reflect"
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
	case reflect.Array:
		err = marshalList(v, w)
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
