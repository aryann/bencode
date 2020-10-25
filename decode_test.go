package bencode

import (
	"reflect"
	"testing"
)

type simpleStruct struct {
	X       int    `key:"x"`
	Y       int    `key:"yy"`
	Z       string `key:"zzz"`
	Unnamed string
}

type compositStruct struct {
	StringList []string       `key:"strings"`
	IntList    []int          `key:"ints"`
	StructList []simpleStruct `key:"structs"`
}

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
		wantErr: "expected terminator for integer at offset 2"},

	{name: "not an integer", in: "iNOT_A_NUMBERe", outputArg: new(int),
		wantErr: "expected integer at offset 1"},

	{name: "unterminated integer", in: "i123", outputArg: new(int),
		wantErr: "expected terminator for integer at offset 4"},

	{name: "incorrectly-terminated integer", in: "i123wrong_terminator", outputArg: new(int),
		wantErr: "expected terminator for integer at offset 4"},

	{name: "empty string 1", in: "0:", outputArg: new(string),
		wantOutput: ""},
	{name: "empty string 2", in: "000:", outputArg: new(string),
		wantOutput: ""},
	{name: "one letter string", in: "1:a", outputArg: new(string),
		wantOutput: "a"},
	{name: "three letter string", in: "3:abc", outputArg: new(string),
		wantOutput: "abc"},

	{name: "extra data string 1", in: "0:abc", outputArg: new(string),
		wantErr: "trailing data at offset 2 cannot be parsed"},
	{name: "extra data string 2", in: "2:abcde", outputArg: new(string),
		wantErr: "trailing data at offset 4 cannot be parsed"},
	{name: "unparsable string length", in: "2x3:abcde", outputArg: new(string),
		wantErr: "expected colon between length and value for string at offset 0"},
	{name: "incorrect length string", in: "100:abc", outputArg: new(string),
		wantErr: "string at offset 0 has length 100, yet there are not that many bytes left"},

	{name: "empty list", in: "le", outputArg: new([]int), wantOutput: *new([]int)},
	{name: "single-element integer list", in: "li651ee", outputArg: new([]int),
		wantOutput: []int{651}},
	{name: "multi-element integer list", in: "li651ee", outputArg: new([]int),
		wantOutput: []int{651}},

	{name: "single-element string list", in: "l3:abce", outputArg: new([]string),
		wantOutput: []string{"abc"}},
	{name: "multi-element string list", in: "l3:abc2:de1:fe", outputArg: new([]string),
		wantOutput: []string{"abc", "de", "f"}},

	{name: "unterminated list 1", in: "li651e", outputArg: new([]int),
		wantErr: "expected terminator for list at offset 6"},
	{name: "unterminated list 2", in: "li651ewrong_terminator", outputArg: new([]int),
		wantErr: "expected start of integer, string, list, or dictionary at offset 6"},
	{name: "unterminated list 3", in: "l3:abc", outputArg: new([]string),
		wantErr: "expected terminator for list at offset 6"},
	{name: "unterminated list 4", in: "l", outputArg: new([]string),
		wantErr: "expected terminator for list at offset 1"},
	{name: "unterminated list item", in: "li651", outputArg: new([]int),
		wantErr: "expected terminator for integer at offset 5"},

	{name: "empty dictionary 1", in: "de", outputArg: new(struct{}),
		wantOutput: struct{}{}},
	{name: "empty dictionary 2", in: "de", outputArg: new(simpleStruct),
		wantOutput: simpleStruct{}},

	{name: "single-entry dictionary", in: "d1:xi651ee", outputArg: new(simpleStruct),
		wantOutput: simpleStruct{X: 651}},
	{name: "multi-entry dictionary", in: "d1:xi651e2:yyi123e3:zzz5:helloe",
		outputArg:  new(simpleStruct),
		wantOutput: simpleStruct{X: 651, Y: 123, Z: "hello"}},

	{name: "strings composit dictionary 1", in: "d7:stringslee",
		outputArg:  new(compositStruct),
		wantOutput: compositStruct{}},
	{name: "strings composit dictionary 2", in: "d7:stringsl1:aee",
		outputArg:  new(compositStruct),
		wantOutput: compositStruct{StringList: []string{"a"}}},
	{name: "strings composit dictionary 3", in: "d7:stringsl5:hello6:world!ee",
		outputArg:  new(compositStruct),
		wantOutput: compositStruct{StringList: []string{"hello", "world!"}}},

	{name: "ints composit dictionary 1", in: "d4:intslee",
		outputArg:  new(compositStruct),
		wantOutput: compositStruct{}},
	{name: "ints composit dictionary 2", in: "d4:intsli651eee",
		outputArg:  new(compositStruct),
		wantOutput: compositStruct{IntList: []int{651}}},
	{name: "ints composit dictionary 3", in: "d4:intsli1ei2ei3eee",
		outputArg:  new(compositStruct),
		wantOutput: compositStruct{IntList: []int{1, 2, 3}}},

	{name: "structs composit dictionary 1", in: "d7:structslee",
		outputArg:  new(compositStruct),
		wantOutput: compositStruct{}},
	{name: "structs composit dictionary 2", in: "d7:structsldeee",
		outputArg:  new(compositStruct),
		wantOutput: compositStruct{StructList: []simpleStruct{{}}}},
	{name: "structs composit dictionary 3", in: "d7:structsldededeee",
		outputArg:  new(compositStruct),
		wantOutput: compositStruct{StructList: []simpleStruct{{}, {}, {}}}},
	{name: "structs composit dictionary 4", in: "d7:structsld1:xi651eeee",
		outputArg:  new(compositStruct),
		wantOutput: compositStruct{StructList: []simpleStruct{{X: 651}}}},
	{name: "structs composit dictionary 5",
		in:         "d7:structsld1:xi651e2:yyi600e3:zzz5:helloeee",
		outputArg:  new(compositStruct),
		wantOutput: compositStruct{StructList: []simpleStruct{{X: 651, Y: 600, Z: "hello"}}}},
	{name: "structs composit dictionary 6",
		in:        "d7:structsld1:xi651e2:yyi600e3:zzz5:helloed1:xi751e2:yyi700e3:zzz7:goodbyeed1:xi851e2:yyi800e3:zzz5:helloeee",
		outputArg: new(compositStruct),
		wantOutput: compositStruct{StructList: []simpleStruct{
			{X: 651, Y: 600, Z: "hello"},
			{X: 751, Y: 700, Z: "goodbye"},
			{X: 851, Y: 800, Z: "hello"},
		}}},

	{name: "unterminated dictionary 1", in: "d", outputArg: new(struct{}),
		wantErr: "expected terminator for dictionary at offset 1"},
}

func TestDecode(t *testing.T) {
	for _, testCase := range decodeTests {
		t.Run(testCase.name, func(t *testing.T) {
			got := reflect.New(reflect.TypeOf(testCase.outputArg).Elem())
			err := Unmarshal(testCase.in, got.Interface())
			if testCase.wantErr != "" || err != nil {
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
