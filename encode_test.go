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

	{name: "empty array", in: [0]string{}, wantOutput: "le"},
	{name: "string array", in: [3]string{"a", "bcd", "efghi"}, wantOutput: "l1:a3:bcd5:efghie"},
	{name: "int array", in: [3]int{1, 234, 5678}, wantOutput: "li1ei234ei5678ee"},
	{name: "mixed-type array", in: [4]interface{}{123, "abc", 456, "def"}, wantOutput: "li123e3:abci456e3:defe"},

	{name: "empty slice", in: []string{}, wantOutput: "le"},
	{name: "string slice", in: []string{"a", "bcd", "efghi"}, wantOutput: "l1:a3:bcd5:efghie"},
	{name: "int slice", in: []int{1, 234, 5678}, wantOutput: "li1ei234ei5678ee"},
	{name: "mixed-type slice", in: []interface{}{123, "abc", 456, "def"}, wantOutput: "li123e3:abci456e3:defe"},

	{name: "empty struct", in: struct{}{}, wantOutput: "de"},

	{name: "single-field struct",
		in: struct {
			x string `bencode:"my-field"`
		}{
			x: "hello",
		},
		wantOutput: "d8:my-field5:helloe"},

	{
		name: "multi-field struct",
		in: struct {
			x string `bencode:"my-field-1"`
			y string `bencode:"my-field-2"`
			z int    `bencode:"my-field-3"`
		}{
			x: "hello",
			y: "world",
			z: 123,
		},
		wantOutput: "d10:my-field-15:hello10:my-field-25:world10:my-field-3i123ee",
	},

	{
		name: "missing tag struct",
		in: struct {
			x string
		}{
			x: "hello",
		},
		wantErr: "found struct field with no 'bencode' tag",
	},

	{
		name: "incorrect tag struct",
		in: struct {
			x string `bad-tag-name:"my-field"`
		}{
			x: "hello",
		},
		wantErr: "found struct field with no 'bencode' tag",
	},

	{
		name: "list-containing struct",
		in: struct {
			stringArray [3]string `bencode:"string-array"`
			stringSlice []string  `bencode:"string-slice"`
		}{
			stringArray: [3]string{"a", "b", "c"},
			stringSlice: []string{"x", "y", "z"},
		},
		wantOutput: "d12:string-arrayl1:a1:b1:ce12:string-slicel1:x1:y1:zee",
	},

	{
		name: "struct-containing struct",
		in: struct {
			structField struct {
				a int `bencode:"a"`
				b int `bencode:"b"`
			} `bencode:"struct"`
			structArray [3]struct {
				c int `bencode:"c"`
			} `bencode:"struct-array"`
			structSlice []struct {
				d int `bencode:"d"`
			} `bencode:"struct-slice"`
		}{
			structField: struct {
				a int `bencode:"a"`
				b int `bencode:"b"`
			}{
				a: 123,
				b: 456,
			},
			structArray: [3]struct {
				c int `bencode:"c"`
			}{{c: 1}, {c: 2}, {c: 3}},
			structSlice: []struct {
				d int `bencode:"d"`
			}{{d: 1}, {d: 2}, {d: 3}},
		},
		wantOutput: "d6:structd1:ai123e1:bi456ee12:struct-arrayld1:ci1eed1:ci2eed1:ci3eee12:struct-sliceld1:di1eed1:di2eed1:di3eeee",
	},

	{
		name: "bencode sorting in struct",
		in: struct {
			c    string `bencode:"c"`
			b    string `bencode:"b"`
			a    string `bencode:"a"`
			zero string `bencode:"0"`
		}{
			c:    "C",
			b:    "B",
			a:    "A",
			zero: "ZERO",
		},
		wantOutput: "d1:04:ZERO1:a1:A1:b1:B1:c1:Ce",
	},
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
			if string(out) != testCase.wantOutput {
				if err == nil {
					t.Errorf("got output '%s', want '%s'", out, testCase.wantOutput)
				} else {
					t.Errorf("got unexpected error: %v", err)
				}
			}
		})
	}
}
