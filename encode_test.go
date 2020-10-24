package bencode

import (
	"testing"
)

var encodeTests = []struct {
	name       string
	in         interface{}
	wantErr    string
	wantOutput string
}{
	{name: "empty string", in: "", wantOutput: "0:"},
	{name: "string", in: "hello", wantOutput: "5:hello"},
	{name: "string with space", in: "Hello, world!", wantOutput: "13:Hello, world!"},
	{name: "string with non-ascii characters", in: "§", wantErr: "strings may not contain non-ascii characters: §"},
}

func TestEncode(t *testing.T) {
	for _, testCase := range encodeTests {
		t.Run(testCase.name, func(t *testing.T) {
			out, err := Marshal(testCase.in)
			if testCase.wantErr != "" {
				if err == nil {
					t.Errorf("want error with message '%v', got no error", testCase.wantErr)
				} else if err.Error() != testCase.wantErr {
					t.Errorf("got error '%v', want '%v'", err, testCase.wantErr)
				}
			}
			if out != testCase.wantOutput {
				t.Errorf("got output '%s', want '%s'", out, testCase.wantOutput)
			}
		})
	}
}
