## Yet another Config Parser for go

This is a config parser most similar to Nginx

[![GoDoc](https://godoc.org/github.com/lytics/confl?status.svg)](https://godoc.org/github.com/lytics/confl).    

Use [SubmlimeText Nginx Plugin](https://github.com/brandonwamboldt/sublime-nginx)

### Examples

This package works similarly to how the Go standard library handles `XML`
and `JSON`. Namely, data is loaded into Go values via reflection.

For the simplest example, consider a file as just a list of keys
and values:

```
// Comments in Config
Age = 25
# another comment
Cats = [ "Cauchy", "Plato" ]
# now, using quotes on key
"Pi" = 3.14
Perfection = [ 6, 28, 496, 8128 ]
DOB = 1987-07-05T05:45:00Z
```

Which could be defined in Go as:

```go
type Config struct {
  Age int
  Cats []string
  Pi float64
  Perfection []int
  DOB time.Time 
}
```

And then decoded with:

```go
var conf Config
if _, err := confl.Decode(data, &conf); err != nil {
  // handle error
}
```

You can also use struct tags if your struct field name doesn't map to a confl
key value directly:

```
some_key_NAME = "wat"
```

```go
type Config struct {
  ObscureKey string `confl:"some_key_NAME"`
}
```

### Using the `encoding.TextUnmarshaler` interface

Here's an example that automatically parses duration strings into 
`time.Duration` values:

```
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
```

Which can be decoded with:

```go
type song struct {
  Name     string
  Duration duration
}
type songs struct {
  Song []song
}
var favorites songs
if _, err := Decode(blob, &favorites); err != nil {
  log.Fatal(err)
}

for _, s := range favorites.Song {
  fmt.Printf("%s (%s)\n", s.Name, s.Duration)
}
```

And you'll also need a `duration` type that satisfies the 
`encoding.TextUnmarshaler` interface:

```go
type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
```

### More complex usage

Here's an example of how to load the example from the official spec page:

```
# nice, config with comments

title = "conf Example"

hand {
  name = "Tyrion"
  organization = "Lannisters"
  bio = "Imp"                 // comments on fields
  dob = 1979-05-27T07:32:00Z  # dates, and more comments on fields
}

// Now, some name/value that is quoted and more json esque
address : {
  "street"  : "1 Sky Cell",
  "city"    : "Eyre",
  "region"  : "Vale of Arryn",
  "country" : "Westeros"
}

servers {
  # You can indent as you please. Tabs or spaces. 
  alpha {
    ip = "10.0.0.1"
    dc = "eqdc10"
  }

  beta {
    ip = "10.0.0.2"
    dc = "eqdc10"
  }

}

clients {
	data = [ ["gamma", "delta"], [1, 2] ] # just an update to make sure parsers support it

	# Line breaks are OK when inside arrays
	hosts = [
	  "alpha",
	  "omega"
	]
}

```

And the corresponding Go types are:

```go
type Config struct {
	Title string
	Owner ownerInfo
	DB database `confl:"database"`
	Servers map[string]server
	Clients clients
}

type ownerInfo struct {
	Name string
	Org string `confl:"organization"`
	Bio string
	DOB time.Time
}

type database struct {
	Server string
	Ports []int
	ConnMax int `confl:"connection_max"`
	Enabled bool
}

type server struct {
	IP string
	DC string
}

type clients struct {
	Data [][]interface{}
	Hosts []string
}
```

Note that a case insensitive match will be tried if an exact match can't be
found.

A working example of the above can be found in `_examples/example.{go,conf}`.

