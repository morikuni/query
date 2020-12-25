package query

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParser(t *testing.T) {
	p := NewParser("&")

	set := NewOpSet("=")
	nameCond := p.String("name", set)
	idCond := p.String("id", set)

	v := url.Values{}
	v.Add("name", "hello")
	v.Add("id", "world")
	err := p.Parse(v.Encode())
	if err != nil {
		t.Fatal(err)
	}

	mustEqual(t, nameCond, &StringCondition{
		Key:   "name",
		Op:    "=",
		Value: "hello",
	})
	mustEqual(t, idCond, &StringCondition{
		Key:   "id",
		Op:    "=",
		Value: "world",
	})
}

func mustEqual(tb testing.TB, got, want interface{}) {
	tb.Helper()

	if diff := cmp.Diff(got, want); diff != "" {
		tb.Fatalf("(-want, +got)\n%s", diff)
	}
}
