package confl

import (
	"fmt"
	"reflect"
	"testing"
)

// Test to make sure we get what we expect.

func test(t *testing.T, data string, ex map[string]interface{}) {
	m, err := Parse(data)
	if err != nil {
		t.Fatalf("Received err: %v\n", err)
	}
	if m == nil {
		t.Fatal("Received nil map")
	}

	if !reflect.DeepEqual(m, ex) {
		t.Fatalf("Not Equal:\nReceived: '%+v'\nExpected: '%+v'\n", m, ex)
	}
}

func TestParseSimpleTopLevel(t *testing.T) {
	ex := map[string]interface{}{
		"foo": "1",
		"bar": float64(2.2),
		"baz": true,
		"boo": int64(22),
	}
	test(t, "foo='1'; bar=2.2; baz=true; boo=22", ex)
}

var sample1 = `
foo  {
  host {
    ip   = '127.0.0.1'
    port = 4242
  }
  servers = [ "a.com", "b.com", "c.com"]
}
`

func TestParseSample1(t *testing.T) {
	ex := map[string]interface{}{
		"foo": map[string]interface{}{
			"host": map[string]interface{}{
				"ip":   "127.0.0.1",
				"port": int64(4242),
			},
			"servers": []interface{}{"a.com", "b.com", "c.com"},
		},
	}
	test(t, sample1, ex)
}

var cluster = `
cluster {
  port: 4244

  authorization {
    user: route_user
    password: top_secret
    timeout: 1
  }

  # Routes are actively solicited and connected to from this server.
  # Other servers can connect to us if they supply the correct credentials
  # in their routes definitions from above.

  // Test both styles of comments

  routes = [
    nats-route://foo:bar@apcera.me:4245
    nats-route://foo:bar@apcera.me:4246
  ]
}
`

func TestParseSample2(t *testing.T) {
	ex := map[string]interface{}{
		"cluster": map[string]interface{}{
			"port": int64(4244),
			"authorization": map[string]interface{}{
				"user":     "route_user",
				"password": "top_secret",
				"timeout":  int64(1),
			},
			"routes": []interface{}{
				"nats-route://foo:bar@apcera.me:4245",
				"nats-route://foo:bar@apcera.me:4246",
			},
		},
	}

	test(t, cluster, ex)
}

var sample3 = `
foo  {
  expr = '(true == "false")'
  text = 'This is a multi-line
text block.'
  text2 (
    hello world
      this is multi line
)
}
`

func TestParseSample3(t *testing.T) {
	ex := map[string]interface{}{
		"foo": map[string]interface{}{
			"expr":  "(true == \"false\")",
			"text":  "This is a multi-line\ntext block.",
			"text2": "hello world\n  this is multi line",
		},
	}
	test(t, sample3, ex)
}

var sample4 = `
  array [
    { abc: 123 }
    { xyz: "word" }
  ]
`

func TestParseSample4(t *testing.T) {
	ex := map[string]interface{}{
		"array": []interface{}{
			map[string]interface{}{"abc": int64(123)},
			map[string]interface{}{"xyz": "word"},
		},
	}
	test(t, sample4, ex)
}

var sample5 = `
  table [
    [ 1, 123  ],
    [ "a", "b", "c"],
  ]
`

func TestParseSample5(t *testing.T) {
	ex := map[string]interface{}{
		"table": []interface{}{
			[]interface{}{int64(1), int64(123)},
			[]interface{}{"a", "b", "c"},
		},
	}
	test(t, sample5, ex)
}

func TestBigSlices(t *testing.T) {
	txt := "Hosts   : ["
	for i := 0; i < 100; i++ {
		txt += fmt.Sprintf(`"http://192.168.1.%d:9999", `, i)
	}
	txt += `"http://123.123.123.123:9999"]` + "\n"

	x := struct{ Hosts []string }{}
	if err := Unmarshal([]byte(txt), &x); err != nil {
		t.Fatalf("error unmarshaling sample: %v", err)
	}
	if len(x.Hosts) != 101 {
		t.Fatalf("%d != 101", len(x.Hosts))
	}
	for i, v := range x.Hosts {
		if i < 100 && v != fmt.Sprintf("http://192.168.1.%d:9999", i) {
			t.Errorf("%d unexpected: %s", i, v)
		}
	}
	if x.Hosts[100] != "http://123.123.123.123:9999" {
		t.Errorf("unexpected: %s", x.Hosts[100])
	}
}
