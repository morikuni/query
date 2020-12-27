package query

import (
	"net/url"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/arbitrary"
	"github.com/leanovate/gopter/gen"

	"github.com/google/go-cmp/cmp"
)

func TestParser(t *testing.T) {
	p := NewParser("\n")

	set := NewOpSet(Equal, NotEqual, LessThan, LessThanOrEqual, GreaterThan, GreaterThanOrEqual)
	str := p.String("str", set)
	strs := p.StringSlice("strs", set)
	number := p.Int64("number", set)
	numbers := p.Int64Slice("numbers", set)
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	ts := p.Timestamp("ts", set, jst)
	b := p.Bool("bool", set)
	f := p.Float64("float", set)

	q := `
str = hello
strs < apple,orange, grape
number<=39
numbers >= 11, 22 ,33
ts >2020-12-26 14:20:33
bool!= true
float=3.14 `
	err = p.Parse(q)
	if err != nil {
		t.Fatal(err)
	}

	mustEqual(t, str, &String{
		Key:   "str",
		Op:    Equal,
		Value: "hello",
	})
	mustEqual(t, strs, &StringSlice{
		Key:   "strs",
		Op:    LessThan,
		Value: []string{"apple", "orange", "grape"},
	})
	mustEqual(t, number, &Int64{
		Key:   "number",
		Op:    LessThanOrEqual,
		Value: 39,
	})
	mustEqual(t, numbers, &Int64Slice{
		Key:   "numbers",
		Op:    GreaterThanOrEqual,
		Value: []int64{11, 22, 33},
	})
	mustEqual(t, ts, &Timestamp{
		Key:      "ts",
		Op:       GreaterThan,
		Value:    time.Date(2020, 12, 26, 14, 20, 33, 0, jst),
		Location: jst,
	})
	mustEqual(t, b, &Bool{
		Key:   "bool",
		Op:    NotEqual,
		Value: true,
	})
	mustEqual(t, f, &Float64{
		Key:   "float",
		Op:    Equal,
		Value: 3.14,
	})
}

func mustEqual(tb testing.TB, got, want interface{}) {
	tb.Helper()

	options := []cmp.Option{
		cmp.Comparer(func(a, b time.Location) bool {
			return a.String() == b.String()
		}),
	}

	if diff := cmp.Diff(want, got, options...); diff != "" {
		tb.Fatalf("(-want, +got)\n%s", diff)
	}
}

func TestDelimiterSplitter(t *testing.T) {
	text := `
a=\"a
 b = b"
"
c = "ccc
ccc\"
ccc "
"d"= d 
`

	s := delimiterSplitter{[]byte("\n")}
	conds := s.Split([]byte(text))

	toStrings := func(bs [][]byte) []string {
		ss := make([]string, len(bs))
		for i := range bs {
			ss[i] = string(bs[i])
		}
		return ss
	}

	mustEqual(t, toStrings(conds), []string{
		"",
		`a=\"a`,
		` b = b"
"`,
		`c = "ccc
ccc\"
ccc "`,
		`"d"= d `,
	})
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

	properties := gopter.NewProperties(nil)

	properties.Property("query parameter", params.ForAll(func(val url.Values) bool {
		p := NewParser("&")
		conds := make([]*String, 0, len(val))
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
