package confl

import (
	"flag"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	u "github.com/araddon/gou"
	"github.com/bmizerany/assert"
)

func init() {
	log.SetFlags(0)
	flag.Parse()
	if testing.Verbose() {
		u.SetupLogging("debug")
		u.SetColorOutput()
	}
}

func TestDecodeSimple(t *testing.T) {
	var simpleConfigString = `
age = 250
ageptr2 = 200
andrew = "gallant"
kait = "brady"
now = 1987-07-05T05:45:00Z 
yesOrNo = true
pi = 3.14
colors = [
	["red", "green", "blue"],
	["cyan", "magenta", "yellow", "black"],
	[pink,brown],
]

my {
  Cats {
    plato = "cat 1"
    cauchy = "cat 2"
  }
}

games [
    {
    	// more comments behind tab
        name "game of thrones"  # we have a comment
        sku  "got"   // another comment, empty line next

    }
    {
        name "settlers of catan"
        // a comment
    }
]

`

	type cats struct {
		Plato  string
		Cauchy string
	}
	type game struct {
		Name string
		Sku  string
	}
	type simpleType struct {
		Age     int
		AgePtr  *int
		AgePtr2 *int
		Colors  [][]string
		Pi      float64
		YesOrNo bool
		Now     time.Time
		Andrew  string
		Kait    string
		My      map[string]cats
		Games   []*game
	}

	var simple simpleType
	_, err := Decode(simpleConfigString, &simple)
	assert.Tf(t, err == nil, "err nil?%v", err)

	now, err := time.Parse("2006-01-02T15:04:05", "1987-07-05T05:45:00")
	if err != nil {
		panic(err)
	}
	age200 := int(200)
	var answer = simpleType{
		Age:     250,
		AgePtr2: &age200,
		Andrew:  "gallant",
		Kait:    "brady",
		Now:     now,
		YesOrNo: true,
		Pi:      3.14,
		Colors: [][]string{
			{"red", "green", "blue"},
			{"cyan", "magenta", "yellow", "black"},
			{"pink", "brown"},
		},
		My: map[string]cats{
			"Cats": cats{Plato: "cat 1", Cauchy: "cat 2"},
		},
		Games: []*game{
			&game{"game of thrones", "got"},
			&game{Name: "settlers of catan"},
		},
	}
	assert.Tf(t, simple.AgePtr == nil, "must have nil ptr")
	if !reflect.DeepEqual(simple, answer) {
		t.Fatalf("Expected\n-----\n%#v\n-----\nbut got\n-----\n%#v\n",
			answer, simple)
	}

	// Now Try decoding using Decoder
	var simpleDec simpleType
	decoder := NewDecoder(strings.NewReader(simpleConfigString))
	err = decoder.Decode(&simpleDec)
	assert.Tf(t, err == nil, "err nil?%v", err)

	assert.Tf(t, simpleDec.AgePtr == nil, "must have nil ptr")
	if !reflect.DeepEqual(simpleDec, answer) {
		t.Fatalf("Expected\n-----\n%#v\n-----\nbut got\n-----\n%#v\n",
			answer, simpleDec)
	}

	// Now Try decoding using Unmarshal
	var simple2 simpleType
	err = Unmarshal([]byte(simpleConfigString), &simple2)
	assert.Tf(t, err == nil, "err nil?%v", err)

	assert.Tf(t, simple2.AgePtr == nil, "must have nil ptr")
	if !reflect.DeepEqual(simple2, answer) {
		t.Fatalf("Expected\n-----\n%#v\n-----\nbut got\n-----\n%#v\n",
			answer, simple2)
	}
}

func TestDecodeEmbedded(t *testing.T) {
	type Dog struct{ Name string }
	type Age int

	tests := map[string]struct {
		input       string
		decodeInto  interface{}
		wantDecoded interface{}
	}{
		"embedded struct": {
			input:       `Name = "milton"`,
			decodeInto:  &struct{ Dog }{},
			wantDecoded: &struct{ Dog }{Dog{"milton"}},
		},
		"embedded non-nil pointer to struct": {
			input:       `Name = "milton"`,
			decodeInto:  &struct{ *Dog }{},
			wantDecoded: &struct{ *Dog }{&Dog{"milton"}},
		},
		"embedded nil pointer to struct": {
			input:       ``,
			decodeInto:  &struct{ *Dog }{},
			wantDecoded: &struct{ *Dog }{nil},
		},
		"embedded int": {
			input:       `Age = -5`,
			decodeInto:  &struct{ Age }{},
			wantDecoded: &struct{ Age }{-5},
		},
	}

	for label, test := range tests {
		_, err := Decode(test.input, test.decodeInto)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(test.wantDecoded, test.decodeInto) {
			t.Errorf("%s: want decoded == %+v, got %+v",
				label, test.wantDecoded, test.decodeInto)
		}
	}
}

func TestDecodeTableArrays(t *testing.T) {
	var tableArrays = `
albums [
	{
		name = "Born to Run"
	    songs [
	      { name = "Jungleland" },
	      { name = "Meeting Across the River" }
		]
	}
	{
		name = "Born in the USA"
  	    songs [
	      { name = "Glory Days" },
	      { name = "Dancing in the Dark" }
	    ]
    }
]`

	type Song struct {
		Name string
	}

	type Album struct {
		Name  string
		Songs []Song
	}

	type Music struct {
		Albums []Album
	}

	expected := Music{[]Album{
		{"Born to Run", []Song{{"Jungleland"}, {"Meeting Across the River"}}},
		{"Born in the USA", []Song{{"Glory Days"}, {"Dancing in the Dark"}}},
	}}
	var got Music
	if _, err := Decode(tableArrays, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("\n%#v\n!=\n%#v\n", expected, got)
	}
}

// Case insensitive matching tests.
// A bit more comprehensive than needed given the current implementation,
// but implementations change.
// Probably still missing demonstrations of some ugly corner cases regarding
// case insensitive matching and multiple fields.
func TestDecodeCase(t *testing.T) {
	var caseData = `
tOpString = "string"
tOpInt = 1
tOpFloat = 1.1
tOpBool = true
tOpdate = 2006-01-02T15:04:05Z
tOparray = [ "array" ]
Match = "i should be in Match only"
MatcH = "i should be in MatcH only"
once = "just once"
nEst {
	eD {
		nEstedString = "another string"
	}
}
`

	type InsensitiveEd struct {
		NestedString string
	}

	type InsensitiveNest struct {
		Ed InsensitiveEd
	}

	type Insensitive struct {
		TopString string
		TopInt    int
		TopFloat  float64
		TopBool   bool
		TopDate   time.Time
		TopArray  []string
		Match     string
		MatcH     string
		Once      string
		OncE      string
		Nest      InsensitiveNest
	}

	tme, err := time.Parse(time.RFC3339, time.RFC3339[:len(time.RFC3339)-5])
	if err != nil {
		panic(err)
	}
	expected := Insensitive{
		TopString: "string",
		TopInt:    1,
		TopFloat:  1.1,
		TopBool:   true,
		TopDate:   tme,
		TopArray:  []string{"array"},
		MatcH:     "i should be in MatcH only",
		Match:     "i should be in Match only",
		Once:      "just once",
		OncE:      "",
		Nest: InsensitiveNest{
			Ed: InsensitiveEd{NestedString: "another string"},
		},
	}
	var got Insensitive
	if _, err := Decode(caseData, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("\n%#v\n!=\n%#v\n", expected, got)
	}
}

func TestDecodePointers(t *testing.T) {
	type Object struct {
		Type        string
		Description string
	}

	type Dict struct {
		NamedObject map[string]*Object
		BaseObject  *Object
		Strptr      *string
		Strptrs     []*string
	}
	s1, s2, s3 := "blah", "abc", "def"
	expected := &Dict{
		Strptr:  &s1,
		Strptrs: []*string{&s2, &s3},
		NamedObject: map[string]*Object{
			"foo": {"FOO", "fooooo!!!"},
			"bar": {"BAR", "ba-ba-ba-ba-barrrr!!!"},
		},
		BaseObject: &Object{"BASE", "da base"},
	}

	ex1 := `
Strptr = "blah"
Strptrs = ["abc", "def"]

NamedObject {
	foo {
      Type = "FOO"
      Description = "fooooo!!!"
	}

    bar {
		Type = "BAR"
		Description = "ba-ba-ba-ba-barrrr!!!"
    }
}

BaseObject {
    Type = "BASE"
    Description = "da base"
}
`
	dict := new(Dict)
	_, err := Decode(ex1, dict)
	if err != nil {
		t.Errorf("Decode error: %v", err)
	}
	if !reflect.DeepEqual(expected, dict) {
		t.Fatalf("\n%#v\n!=\n%#v\n", expected, dict)
	}
}

type sphere struct {
	Center [3]float64
	Radius float64
}

func TestDecodeSimpleArray(t *testing.T) {
	var s1 sphere
	if _, err := Decode(`center = [0.0, 1.5, 0.0]`, &s1); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeArrayWrongSize(t *testing.T) {
	var s1 sphere
	if _, err := Decode(`center = [0.1, 2.3]`, &s1); err == nil {
		t.Fatal("Expected array type mismatch error")
	}
}

func TestDecodeLargeIntoSmallInt(t *testing.T) {
	type table struct {
		Value int8
	}
	var tab table
	if _, err := Decode(`value = 500`, &tab); err == nil {
		t.Fatal("Expected integer out-of-bounds error.")
	}
}

func TestDecodeSizedInts(t *testing.T) {
	type table struct {
		U8  uint8
		U16 uint16
		U32 uint32
		U64 uint64
		U   uint
		I8  int8
		I16 int16
		I32 int32
		I64 int64
		I   int
	}
	answer := table{1, 1, 1, 1, 1, -1, -1, -1, -1, -1}
	configStr := `
	u8 = 1
	u16 = 1
	u32 = 1
	u64 = 1
	u = 1
	i8 = -1
	i16 = -1
	i32 = -1
	i64 = -1
	i = -1
	`
	var tab table
	if _, err := Decode(configStr, &tab); err != nil {
		t.Fatal(err.Error())
	}
	if answer != tab {
		t.Fatalf("Expected %#v but got %#v", answer, tab)
	}
}

func ExampleMetaData_PrimitiveDecode() {
	var md MetaData
	var err error

	var rawData = `

ranking = ["Springsteen", "JGeils"]

bands {
	Springsteen { 
		started = 1973
		albums = ["Greetings", "WIESS", "Born to Run", "Darkness"]
	}

	JGeils {
		started = 1970
		albums = ["The J. Geils Band", "Full House", "Blow Your Face Out"]
	}
}
`

	type band struct {
		Started int
		Albums  []string
	}
	type classics struct {
		Ranking []string
		Bands   map[string]Primitive
	}

	// Do the initial decode. Reflection is delayed on Primitive values.
	var music classics
	if md, err = Decode(rawData, &music); err != nil {
		log.Fatal(err)
	}

	// MetaData still includes information on Primitive values.
	fmt.Printf("Is `bands.Springsteen` defined? %v\n",
		md.IsDefined("bands", "Springsteen"))

	// Decode primitive data into Go values.
	for _, artist := range music.Ranking {
		// A band is a primitive value, so we need to decode it to get a
		// real `band` value.
		primValue := music.Bands[artist]

		var aBand band
		if err = md.PrimitiveDecode(primValue, &aBand); err != nil {
			u.Warnf("failed on: ")
			log.Fatalf("Failed on %v  %v", primValue, err)
		}
		fmt.Printf("%s started in %d.\n", artist, aBand.Started)
	}
	// Check to see if there were any fields left undecoded.
	// Note that this won't be empty before decoding the Primitive value!
	fmt.Printf("Undecoded: %q\n", md.Undecoded())

	// Output:
	// Is `bands.Springsteen` defined? true
	// Springsteen started in 1973.
	// JGeils started in 1970.
	// Undecoded: []
}

func ExampleDecode() {
	var rawData = `
# Some comments.
alpha {
	ip = "10.0.0.1"

	config {
		Ports = [ 8001, 8002 ]
		Location = "Toronto"
		Created = 1987-07-05T05:45:00Z
	}
}


beta {
	ip = "10.0.0.2"

	config {
		Ports = [ 9001, 9002 ]
		Location = "New Jersey"
		Created = 1887-01-05T05:55:00Z
	}
}

`

	type serverConfig struct {
		Ports    []int
		Location string
		Created  time.Time
	}

	type server struct {
		IP     string       `confl:"ip"`
		Config serverConfig `confl:"config"`
	}

	type servers map[string]server

	var config servers
	if _, err := Decode(rawData, &config); err != nil {
		log.Fatal(err)
	}

	for _, name := range []string{"alpha", "beta"} {
		s := config[name]
		fmt.Printf("Server: %s (ip: %s) in %s created on %s\n",
			name, s.IP, s.Config.Location,
			s.Config.Created.Format("2006-01-02"))
		fmt.Printf("Ports: %v\n", s.Config.Ports)
	}

	// Output:
	// Server: alpha (ip: 10.0.0.1) in Toronto created on 1987-07-05
	// Ports: [8001 8002]
	// Server: beta (ip: 10.0.0.2) in New Jersey created on 1887-01-05
	// Ports: [9001 9002]
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// Example Unmarshaler shows how to decode strings into your own
// custom data type.
func Example_unmarshaler() {
	rawData := `
song [
	{
		name = "Thunder Road"
		duration = "4m49s"
	},
	{
		name = "Stairway to Heaven"
		duration = "8m03s"
	}
]

`
	type song struct {
		Name     string
		Duration duration
	}
	type songs struct {
		Song []song
	}
	var favorites songs
	if _, err := Decode(rawData, &favorites); err != nil {
		log.Fatal(err)
	}

	// Code to implement the TextUnmarshaler interface for `duration`:
	//
	// type duration struct {
	// 	time.Duration
	// }
	//
	// func (d *duration) UnmarshalText(text []byte) error {
	// 	var err error
	// 	d.Duration, err = time.ParseDuration(string(text))
	// 	return err
	// }

	for _, s := range favorites.Song {
		fmt.Printf("%s (%s)\n", s.Name, s.Duration)
	}
	// Output:
	// Thunder Road (4m49s)
	// Stairway to Heaven (8m3s)
}

// Example StrictDecoding shows how to detect whether there are keys in the
// config document that weren't decoded into the value given. This is useful
// for returning an error to the user if they've included extraneous fields
// in their configuration.
// func Example_strictDecoding() {
// 	var rawData = `
// key1 = "value1"
// key2 = "value2"
// key3 = "value3"
// `
// 	type config struct {
// 		Key1 string
// 		Key3 string
// 	}

// 	var conf config
// 	md, err := Decode(rawData, &conf)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("Undecoded keys: %q\n", md.Undecoded())
// 	// Output:
// 	// Undecoded keys: ["key2"]
// }
