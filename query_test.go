package query

import (
	"net/url"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/arbitrary"
	"github.com/leanovate/gopter/gen"

	"github.com/google/go-cmp/cmp"
)

func TestParser(t *testing.T) {
	p := NewParser("&")

	set := NewOpSet("=")
	nameCond := p.String("name", set)
	idCond := p.String("id", set)
	likeCond := p.StringSlice("like", set)
	ageCond := p.Int64("age", set)

	v := url.Values{}
	v.Add("name", "hello")
	v.Add("id", "world")
	v.Add("like", "apple,orange, grape")
	v.Add("age", "39")
	q, err := url.QueryUnescape(v.Encode())
	if err != nil {
		t.Fatal(err)
	}
	err = p.Parse(q)
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
	mustEqual(t, likeCond, &StringSliceCondition{
		Key:   "like",
		Op:    "=",
		Value: []string{"apple", "orange", "grape"},
	})
	mustEqual(t, ageCond, &Int64Condition{
		Key:   "age",
		Op:    "=",
		Value: 39,
	})
}

func mustEqual(tb testing.TB, got, want interface{}) {
	tb.Helper()

	if diff := cmp.Diff(want, got); diff != "" {
		tb.Fatalf("(-want, +got)\n%s", diff)
	}
}

func TestParserPBT(t *testing.T) {
	params := arbitrary.DefaultArbitraries()
	nonEmptyString := gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	})
	params.RegisterGen(gen.MapOf(nonEmptyString, gen.AlphaString()).Map(func(m map[string]string) url.Values {
		val := url.Values{}
		for k, v := range m {
			val.Set(k, v)
		}
		return val
	}))

	setting := gopter.DefaultTestParameters()
	setting.MinSuccessfulTests = 1000
	properties := gopter.NewProperties(setting)

	properties.Property("query parameter", params.ForAll(func(val url.Values) bool {
		p := NewParser("&")
		conds := make([]*StringCondition, 0, len(val))
		for k := range val {
			conds = append(conds, p.String(k, NewOpSet(Equal)))
		}

		err := p.Parse(val.Encode())
		if err != nil {
			return false
		}

		for _, c := range conds {
			if c.Op != Equal {
				return false
			}
			if c.Value != val.Get(c.Key) {
				return false
			}
		}

		return true
	}))

	properties.TestingRun(t)
}
