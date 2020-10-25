package bencode

import (
	"reflect"
	"testing"
)

var decodeTests = []struct {
	name       string
	in         string
	outputArg  interface{}
	wantErr    string
	wantOutput interface{}
}{
	{name: "empty input", in: "", outputArg: new(string),
		wantErr: "no data to read at offset 0"},

	{name: "zero integer", in: "i0e", outputArg: new(int), wantOutput: 0},
	{name: "positive integer", in: "i651e", outputArg: new(int), wantOutput: 651},
	{name: "negative integer", in: "i-601e", outputArg: new(int), wantOutput: -601},

	{name: "missing integer", in: "ie", outputArg: new(int),
		wantErr: "expected integer at offset 1"},

	{name: "malformed integer 1", in: "i-e", outputArg: new(int),
		wantErr: "expected integer at offset 1"},

	{name: "malformed integer 2", in: "i*e", outputArg: new(int),
		wantErr: "expected integer at offset 1"},

	{name: "malformed integer 3", in: "i0x80e", outputArg: new(int),
		wantErr: "expected terminator 'e' for integer at offset 0"},

	{name: "not an integer", in: "iNOT_A_NUMBERe", outputArg: new(int),
		wantErr: "expected integer at offset 1"},

	{name: "unterminated integer", in: "i123", outputArg: new(int),
		wantErr: "expected integer termination at offset 4"},

	{name: "incorrectly-terminated integer", in: "i123wrong_terminator", outputArg: new(int),
		wantErr: "expected terminator 'e' for integer at offset 0"},
}

func TestDecode(t *testing.T) {
	for _, testCase := range decodeTests {
		t.Run(testCase.name, func(t *testing.T) {
			got := reflect.New(reflect.TypeOf(testCase.outputArg).Elem())
			err := Unmarshal(testCase.in, got.Interface())
			if testCase.wantErr != "" {
				if err == nil {
					t.Errorf("want error with message '%v', got no error", testCase.wantErr)
				} else if err.Error() != testCase.wantErr {
					t.Errorf("got error '%v', want '%v'", err, testCase.wantErr)
				}
				return
			}
			if !reflect.DeepEqual(got.Elem().Interface(), testCase.wantOutput) {
				if err == nil {
					t.Errorf("got output '%+v', want '%+v'", got.Elem().Interface(), testCase.wantOutput)
				} else {
					t.Errorf("got unexpected error: %v", err)
				}
			}
		})
	}
}
