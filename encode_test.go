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
	{name: "positive int", in: int(123), wantOutput: "i123e"},
	{name: "zero int", in: int(0), wantOutput: "i0e"},
	{name: "negative int", in: int(-123), wantOutput: "i-123e"},

	{name: "positive int8", in: int8(123), wantOutput: "i123e"},
	{name: "zero int8", in: int8(0), wantOutput: "i0e"},
	{name: "negative int8", in: int8(-123), wantOutput: "i-123e"},

	{name: "positive int16", in: int16(123), wantOutput: "i123e"},
	{name: "zero int16", in: int16(0), wantOutput: "i0e"},
	{name: "negative int16", in: int16(-123), wantOutput: "i-123e"},

	{name: "positive int32", in: int32(123), wantOutput: "i123e"},
	{name: "zero int32", in: int32(0), wantOutput: "i0e"},
	{name: "negative int32", in: int32(-123), wantOutput: "i-123e"},

	{name: "positive int64", in: int64(123), wantOutput: "i123e"},
	{name: "zero int64", in: int64(0), wantOutput: "i0e"},
	{name: "negative int64", in: int64(-123), wantOutput: "i-123e"},

	{name: "positive uint", in: uint(123), wantOutput: "i123e"},
	{name: "zero uint", in: uint(0), wantOutput: "i0e"},

	{name: "positive uint8", in: uint8(123), wantOutput: "i123e"},
	{name: "zero uint8", in: uint8(0), wantOutput: "i0e"},

	{name: "positive uint16", in: uint16(123), wantOutput: "i123e"},
	{name: "zero uint16", in: uint16(0), wantOutput: "i0e"},

	{name: "positive uint32", in: uint32(123), wantOutput: "i123e"},
	{name: "zero uint32", in: uint32(0), wantOutput: "i0e"},

	{name: "positive int64", in: uint64(123), wantOutput: "i123e"},
	{name: "zero int64", in: uint64(0), wantOutput: "i0e"},

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
