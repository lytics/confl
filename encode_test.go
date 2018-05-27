package confl

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	u "github.com/araddon/gou"
	"github.com/stretchr/testify/assert"
)

var _ = u.EMPTY

func TestEncodeRoundTrip(t *testing.T) {
	type Config struct {
		Age        int
		Cats       []string
		Pi         float64
		Perfection []int
		DOB        time.Time
		Ipaddress  net.IP
	}

	var inputs = Config{
		13,
		[]string{"one", "two", "three"},
		3.145,
		[]int{11, 2, 3, 4},
		time.Now(),
		net.ParseIP("192.168.59.254"),
	}

	var firstBuffer bytes.Buffer
	e := NewEncoder(&firstBuffer)
	err := e.Encode(inputs)
	if err != nil {
		t.Fatal(err)
	}
	var outputs Config
	if _, err := Decode(firstBuffer.String(), &outputs); err != nil {
		log.Printf("Could not decode:\n-----\n%s\n-----\n",
			firstBuffer.String())
		t.Fatal(err)
	}

	// could test each value individually, but I'm lazy
	var secondBuffer bytes.Buffer
	e2 := NewEncoder(&secondBuffer)
	err = e2.Encode(outputs)
	if err != nil {
		t.Fatal(err)
	}
	if firstBuffer.String() != secondBuffer.String() {
		t.Error(
			firstBuffer.String(),
			"\n\n is not identical to\n\n",
			secondBuffer.String())
	}

	// now try with Marshal
	marshalBytes, err := Marshal(&inputs)
	assert.Equal(t, nil, err)
	assert.True(t, bytes.Equal(marshalBytes, firstBuffer.Bytes()), "equal?")
}

func TestEncodeMany(t *testing.T) {
	type Embedded struct {
		Int int `confl:"_int"`
	}
	type NonStruct int

	date := time.Date(2014, 5, 11, 20, 30, 40, 0, time.FixedZone("IST", 3600))
	dateStr := "2014-05-11T19:30:40Z"

	tests := []struct {
		label      string
		input      interface{}
		wantOutput string
		wantError  error
	}{
		{label: "bool field",
			input: struct {
				BoolTrue  bool
				BoolFalse bool
			}{true, false},
			wantOutput: "BoolTrue = true\nBoolFalse = false\n",
		},
		{label: "int fields",
			input: struct {
				Int   int
				Int8  int8
				Int16 int16
				Int32 int32
				Int64 int64
			}{1, 2, 3, 4, 5},
			wantOutput: "Int = 1\nInt8 = 2\nInt16 = 3\nInt32 = 4\nInt64 = 5\n",
		},
		{label: "uint fields",
			input: struct {
				Uint   uint
				Uint8  uint8
				Uint16 uint16
				Uint32 uint32
				Uint64 uint64
			}{1, 2, 3, 4, 5},
			wantOutput: "Uint = 1\nUint8 = 2\nUint16 = 3\nUint32 = 4" +
				"\nUint64 = 5\n",
		},
		{label: "float fields",
			input: struct {
				Float32 float32
				Float64 float64
			}{1.5, 2.5},
			wantOutput: "Float32 = 1.5\nFloat64 = 2.5\n",
		},
		{label: "string field",
			input:      struct{ String string }{"foo"},
			wantOutput: "String = \"foo\"\n",
		},
		{label: "string field and unexported field",
			input: struct {
				String     string
				unexported int
			}{"foo", 0},
			wantOutput: "String = \"foo\"\n",
		},
		{label: "datetime field in UTC",
			input:      struct{ Date time.Time }{date},
			wantOutput: fmt.Sprintf("Date = %s\n", dateStr),
		},
		{label: "datetime field as primitive",
			// Using a map here to fail if isStructOrMap() returns true for
			// time.Time.
			input: map[string]interface{}{
				"Date": date,
				"Int":  1,
			},
			wantOutput: fmt.Sprintf("Date = %s\nInt = 1\n", dateStr),
		},
		{label: "array fields",
			input: struct {
				IntArray0 [0]int
				IntArray3 [3]int
			}{[0]int{}, [3]int{1, 2, 3}},
			wantOutput: "IntArray0 = []\nIntArray3 = [1, 2, 3]\n",
		},
		{label: "slice fields",
			input: struct{ IntSliceNil, IntSlice0, IntSlice3 []int }{
				nil, []int{}, []int{1, 2, 3},
			},
			wantOutput: "IntSlice0 = []\nIntSlice3 = [1, 2, 3]\n",
		},
		{label: "datetime slices",
			input: struct{ DatetimeSlice []time.Time }{
				[]time.Time{date, date},
			},
			wantOutput: fmt.Sprintf("DatetimeSlice = [%s, %s]\n",
				dateStr, dateStr),
		},
		{label: "nested arrays and slices",
			input: struct {
				SliceOfArrays         [][2]int
				ArrayOfSlices         [2][]int
				SliceOfArraysOfSlices [][2][]int
				ArrayOfSlicesOfArrays [2][][2]int
				SliceOfMixedArrays    [][2]interface{}
				ArrayOfMixedSlices    [2][]interface{}
			}{
				[][2]int{{1, 2}, {3, 4}},
				[2][]int{{1, 2}, {3, 4}},
				[][2][]int{
					{
						{1, 2}, {3, 4},
					},
					{
						{5, 6}, {7, 8},
					},
				},
				[2][][2]int{
					{
						{1, 2}, {3, 4},
					},
					{
						{5, 6}, {7, 8},
					},
				},
				[][2]interface{}{
					{1, 2}, {"a", "b"},
				},
				[2][]interface{}{
					{1, 2}, {"a", "b"},
				},
			},
			wantOutput: `SliceOfArrays = [[1, 2], [3, 4]]
ArrayOfSlices = [[1, 2], [3, 4]]
SliceOfArraysOfSlices = [[[1, 2], [3, 4]], [[5, 6], [7, 8]]]
ArrayOfSlicesOfArrays = [[[1, 2], [3, 4]], [[5, 6], [7, 8]]]
SliceOfMixedArrays = [[1, 2], ["a", "b"]]
ArrayOfMixedSlices = [[1, 2], ["a", "b"]]
`,
		},
		{label: "empty slice",
			input:      struct{ Empty []interface{} }{[]interface{}{}},
			wantOutput: "Empty = []\n",
		},
		{label: "(error) slice with element type mismatch (string and integer)",
			input:     struct{ Mixed []interface{} }{[]interface{}{1, "a"}},
			wantError: errArrayMixedElementTypes,
		},
		{label: "(error) slice with element type mismatch (integer and float)",
			input:     struct{ Mixed []interface{} }{[]interface{}{1, 2.5}},
			wantError: errArrayMixedElementTypes,
		},
		{label: "slice with elems of differing Go types, same types",
			input: struct {
				MixedInts   []interface{}
				MixedFloats []interface{}
			}{
				[]interface{}{
					int(1), int8(2), int16(3), int32(4), int64(5),
					uint(1), uint8(2), uint16(3), uint32(4), uint64(5),
				},
				[]interface{}{float32(1.5), float64(2.5)},
			},
			wantOutput: "MixedInts = [1, 2, 3, 4, 5, 1, 2, 3, 4, 5]\n" +
				"MixedFloats = [1.5, 2.5]\n",
		},
		{label: "(error) slice w/ element type mismatch (one is nested array)",
			input: struct{ Mixed []interface{} }{
				[]interface{}{1, []interface{}{2}},
			},
			wantError: errArrayMixedElementTypes,
		},
		{label: "(error) slice with 1 nil element",
			input:     struct{ NilElement1 []interface{} }{[]interface{}{nil}},
			wantError: errArrayNilElement,
		},
		{label: "(error) slice with 1 nil element (and other non-nil elements)",
			input: struct{ NilElement []interface{} }{
				[]interface{}{1, nil},
			},
			wantError: errArrayNilElement,
		},
		{label: "simple map",
			input:      map[string]int{"a": 1, "b": 2},
			wantOutput: "a = 1\nb = 2\n",
		},
		{label: "map with interface{} value type",
			input:      map[string]interface{}{"a": 1, "b": "c"},
			wantOutput: "a = 1\nb = \"c\"\n",
		},
		{label: "map with interface{} value type, some of which are structs",
			input: map[string]interface{}{
				"a": struct{ Int int }{2},
				"b": 1,
			},
			wantOutput: "b = 1\na {\n  Int = 2\n}\n",
		},
		{label: "nested map",
			input: map[string]map[string]int{
				"a": {"b": 1},
				"c": {"d": 2},
			},
			wantOutput: "a {\n  b = 1\n}\nc {\n  d = 2\n}\n",
		},
		{label: "nested struct",
			input: struct{ Struct struct{ Int int } }{
				struct{ Int int }{1},
			},
			wantOutput: "Struct {\n  Int = 1\n}\n",
		},
		{label: "nested struct and non-struct field",
			input: struct {
				Struct struct{ Int int }
				Bool   bool
			}{struct{ Int int }{1}, true},
			wantOutput: "Bool = true\nStruct {\n  Int = 1\n}\n",
		},
		{label: "2 nested structs",
			input: struct{ Struct1, Struct2 struct{ Int int } }{
				struct{ Int int }{1}, struct{ Int int }{2},
			},
			wantOutput: "Struct1 {\n  Int = 1\n}\nStruct2 {\n  Int = 2\n}\n",
		},
		{label: "deeply nested structs",
			input: struct {
				Struct1, Struct2 struct{ Struct3 *struct{ Int int } }
			}{
				struct{ Struct3 *struct{ Int int } }{&struct{ Int int }{1}},
				struct{ Struct3 *struct{ Int int } }{nil},
			},
			wantOutput: "Struct1 {\n  Struct3 {\n    Int = 1\n  }\n}\nStruct2 {\n}\n",
		},
		{label: "nested struct with nil struct elem",
			input: struct {
				Struct struct{ Inner *struct{ Int int } }
			}{
				struct{ Inner *struct{ Int int } }{nil},
			},
			wantOutput: "Struct {\n}\n",
		},
		{label: "nested struct with no fields",
			input: struct {
				Struct struct{ Inner struct{} }
			}{
				struct{ Inner struct{} }{struct{}{}},
			},
			wantOutput: "Struct {\n  Inner {\n  }\n}\n",
		},
		{label: "struct with tags",
			input: struct {
				Struct struct {
					Int int `json:"_int"`
				} `confl:"_struct"`
				Bool bool `confl:"_bool"`
			}{
				struct {
					Int int `json:"_int"`
				}{1}, true,
			},
			wantOutput: "_bool = true\n_struct {\n  _int = 1\n}\n",
		},
		{label: "embedded struct",
			input:      struct{ Embedded }{Embedded{1}},
			wantOutput: "_int = 1\n",
		},
		{label: "embedded *struct",
			input:      struct{ *Embedded }{&Embedded{1}},
			wantOutput: "_int = 1\n",
		},
		{label: "nested embedded struct",
			input: struct {
				Struct struct{ Embedded } `confl:"_struct"`
			}{struct{ Embedded }{Embedded{1}}},
			wantOutput: "_struct {\n  _int = 1\n}\n",
		},
		{label: "nested embedded *struct",
			input: struct {
				Struct struct{ *Embedded } `confl:"_struct"`
			}{struct{ *Embedded }{&Embedded{1}}},
			wantOutput: "_struct {\n  _int = 1\n}\n",
		},
		{label: "array of tables",
			input: struct {
				Structs []*struct{ Int int } `confl:"struct"`
			}{
				[]*struct{ Int int }{{1}, {3}},
			},
			wantOutput: "struct = [\n  {\n    Int = 1\n  },\n  {\n    Int = 3\n  }\n]\n",
		},
		{label: "array of tables order",
			input: map[string]interface{}{
				"map": map[string]interface{}{
					"zero": 5,
					"arr": []map[string]int{
						map[string]int{
							"friend": 5,
						},
					},
				},
			},
			wantOutput: "map {\n  zero = 5\n  arr = [\n    {\n      friend = 5\n    }\n  ]\n}\n",
		},
		{label: "(error) top-level slice",
			input:     []struct{ Int int }{{1}, {2}, {3}},
			wantError: errNoKey,
		},
		{label: "(error) slice of slice",
			input: struct {
				Slices [][]struct{ Int int }
			}{
				[][]struct{ Int int }{{{1}}, {{2}}, {{3}}},
			},
			wantError: errArrayNoTable,
		},
		{label: "(error) map no string key",
			input:     map[int]string{1: ""},
			wantError: errNonString,
		},
		{label: "(error) anonymous non-struct",
			input:     struct{ NonStruct }{5},
			wantError: errAnonNonStruct,
		},
		{label: "(error) empty key name",
			input:     map[string]int{"": 1},
			wantError: errAnything,
		},
		{label: "(error) empty map name",
			input: map[string]interface{}{
				"": map[string]int{"v": 1},
			},
			wantError: errAnything,
		},
	}
	for idx, test := range tests {
		u.Debugf("starting test:  #%d %v", idx, test.label)
		encodeExpected(t, fmt.Sprintf("#%d: %s", idx, test.label), test.input,
			test.wantOutput, test.wantError)
	}
}

func TestEncodeNestedTableArrays(t *testing.T) {
	type song struct {
		Name string `confl:"name"`
	}
	type album struct {
		Name  string `confl:"name"`
		Songs []song `confl:"songs"`
	}
	type springsteen struct {
		Albums []album `confl:"albums"`
	}
	value := springsteen{
		[]album{
			{"Born to Run",
				[]song{{"Jungleland"}, {"Meeting Across the River"}}},
			{"Born in the USA",
				[]song{{"Glory Days"}, {"Dancing in the Dark"}}},
		},
	}
	expected := `albums = [
  {
    name = "Born to Run"
    songs = [
      {
        name = "Jungleland"
      },
      {
        name = "Meeting Across the River"
      }
    ]
  },
  {
    name = "Born in the USA"
    songs = [
      {
        name = "Glory Days"
      },
      {
        name = "Dancing in the Dark"
      }
    ]
  }
]
`
	encodeExpected(t, "nested table arrays", value, expected, nil)
}

func TestEncodeArrayHashWithNormalHashOrder(t *testing.T) {
	type Alpha struct {
		V int
	}
	type Beta struct {
		V int
	}
	type Conf struct {
		V int
		A Alpha
		B []Beta
	}

	val := Conf{
		V: 1,
		A: Alpha{2},
		B: []Beta{{3}},
	}
	expected := "V = 1\nA {\n  V = 2\n}\nB = [\n  {\n    V = 3\n  }\n]\n"
	encodeExpected(t, "array hash with normal hash order", val, expected, nil)
}

func encodeExpected(
	t *testing.T, label string, val interface{}, wantStr string, wantErr error,
) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	err := enc.Encode(val)
	if err != wantErr {
		if wantErr != nil {
			if wantErr == errAnything && err != nil {
				return
			}
			t.Errorf("%s: want Encode error %v, got %v", label, wantErr, err)
		} else {
			t.Errorf("%s: Encode failed: %s", label, err)
		}
	}
	if err != nil {
		return
	}
	if got := buf.String(); wantStr != got {
		u.Debugf("\n\n%s wanted: \n%s\ngot: \n%s", label, wantStr, got)
		for pos, r := range wantStr {
			if len(got)-1 <= pos {
				u.Warnf("len mismatch? %v vs %v", len(got), len(wantStr))
			} else if r != rune(got[pos]) {
				u.Warnf("mismatch at position: %v   %s!=%s", pos, string(r), string(got[pos]))
				break
			}
		}
		t.Fatalf("%s: want\n-----\n%q\n-----\nbut got\n-----\n%q\n-----\n", label, wantStr, got)
	}
}

func ExampleEncoder_Encode() {
	date, _ := time.Parse(time.RFC822, "14 Mar 10 18:00 UTC")
	var config = map[string]interface{}{
		"date":   date,
		"counts": []int{1, 1, 2, 3, 5, 8},
		"hash": map[string]string{
			"key1": "val1",
			"key2": "val2",
		},
	}
	buf := new(bytes.Buffer)
	if err := NewEncoder(buf).Encode(config); err != nil {
		u.Errorf("could not encode: %v", err)
		log.Fatal(err)
	}
	fmt.Println(buf.String())

	// Output:
	// counts = [1, 1, 2, 3, 5, 8]
	// date = 2010-03-14T18:00:00Z
	// hash {
	//   key1 = "val1"
	//   key2 = "val2"
	// }
}
