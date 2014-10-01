package confl

import (
	//"github.com/bmizerany/assert"
	"testing"
)

// Test to make sure we get what we expect.
func expect(t *testing.T, lx *lexer, items []item) {
	for i := 0; i < len(items); i++ {
		item := lx.nextItem()
		if item.typ == itemEOF {
			break
		} else if item.typ == itemError {
			t.Fatal(item.val)
		}
		if item != items[i] {
			t.Fatalf("Testing: '%s'\nExpected %q, received %q\n",
				lx.input, items[i], item)
		}
	}
}

func TestLexSimpleKeyStringValues(t *testing.T) {
	expectedItems := []item{
		{itemKey, "foo", 1},
		{itemString, "bar", 1},
		{itemEOF, "", 1},
	}
	// Double quotes
	lx := lex("foo = \"bar\"")
	expect(t, lx, expectedItems)
	// Single quotes
	lx = lex("foo = 'bar'")
	expect(t, lx, expectedItems)
	// No spaces
	lx = lex("foo='bar'")
	expect(t, lx, expectedItems)
	// NL
	lx = lex("foo='bar'\r\n")
	expect(t, lx, expectedItems)
}

func TestLexSimpleKeyIntegerValues(t *testing.T) {
	expectedItems := []item{
		{itemKey, "foo", 1},
		{itemInteger, "123", 1},
		{itemEOF, "", 1},
	}
	lx := lex("foo = 123")
	expect(t, lx, expectedItems)
	lx = lex("foo=123")
	expect(t, lx, expectedItems)
	lx = lex("foo=123\r\n")
	expect(t, lx, expectedItems)
}
