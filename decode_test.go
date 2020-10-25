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
	{name: "zero integer", in: "i0e", outputArg: new(int), wantOutput: 0},
	{name: "positive integer", in: "i651e", outputArg: new(int), wantOutput: 651},
	{name: "negative integer", in: "i-601e", outputArg: new(int), wantOutput: -601},

	{name: "not an integer", in: "iNOT_A_NUMBERe", outputArg: new(int),
		wantErr: "could not parse integer at offset 0: NOT_A_NUMBER"},
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
