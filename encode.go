package bencode

import (
	"fmt"
	"reflect"
	"unicode"
)

// Marshal returns a bencode encoding of v.
func Marshal(v interface{}) (string, error) {
	kind := reflect.TypeOf(v).Kind()
	value := reflect.ValueOf(v)
	switch kind {
	case reflect.String:
		return marshalString(value.String())
	default:
		return "", fmt.Errorf("encountered unsupported type: %s", kind.String())
	}
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func marshalString(s string) (string, error) {
	if !isASCII(s) {
		return "", fmt.Errorf("strings may not contain non-ascii characters: %s", s)
	}
	return fmt.Sprintf("%d:%s", len(s), s), nil
}
