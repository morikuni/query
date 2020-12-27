package query

import (
	"bytes"
	"strconv"
	"time"
	"unicode"
)

type Op string

type OpSet [][]byte

func NewOpSet(ops ...Op) OpSet {
	s := make([][]byte, 0, len(ops))
	for _, op := range ops {
		s = append(s, []byte(op))
	}
	return s
}

const (
	Equal              = Op("=")
	NotEqual           = Op("!=")
	LessThan           = Op("<")
	LessThanOrEqual    = Op("<=")
	GreaterThan        = Op(">")
	GreaterThanOrEqual = Op(">=")
)

type Parser struct {
	splitter   splitter
	conditions []parseCondition
}

type parseCondition struct {
	key []byte
	set OpSet
	c   Condition
}

func NewParser(delimiter string) *Parser {
	return &Parser{
		delimiterSplitter{[]byte(delimiter)},
		nil,
	}
}

func (p *Parser) Parse(query string) error {
	text := []byte(query)
	conds := p.splitter.Split(text)

	for _, cond := range conds {
		err := p.scanCondition(cond)
		if err != nil {
			return err
		}
	}

	return nil
}

type splitter interface {
	Split(text []byte) [][]byte
}

type delimiterSplitter struct {
	delimiter []byte
}

func (s delimiterSplitter) Split(text []byte) (conds [][]byte) {
	for len(text) != 0 {
		idx := bytes.Index(text, s.delimiter)
		if idx == -1 {
			conds = append(conds, text)
			return conds
		}

		var inQuote bool
		for i := 0; i < len(text); i++ {
			if i == idx && !inQuote {
				conds = append(conds, text[:i])
				text = text[i+1:]
				break
			}
			switch text[i] {
			case '"':
				inQuote = !inQuote
				if !inQuote {
					idx = bytes.Index(text[i:], s.delimiter)
					if idx == -1 {
						conds = append(conds, text)
						return conds
					} else {
						idx += i
					}
				}
			case '\\':
				i++
			}
		}
	}
	return conds
}

func (p *Parser) scanCondition(c []byte) error {
	c = bytes.TrimSpace(c)
	for _, cond := range p.conditions {
		c := c
		if !bytes.HasPrefix(c, cond.key) {
			continue
		}

		key := cond.key
		c = bytes.TrimLeftFunc(c[len(key):], unicode.IsSpace)
		for _, op := range cond.set {
			if !bytes.HasPrefix(c, op) {
				continue
			}

			c = bytes.TrimLeftFunc(c[len(op):], unicode.IsSpace)
			err := cond.c.Set(string(key), Op(op), c)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Parser) Condition(key string, set OpSet, c Condition) {
	p.conditions = append(p.conditions, parseCondition{[]byte(key), set, c})
}

func (p *Parser) String(key string, set OpSet) *String {
	c := new(String)
	p.Condition(key, set, c)
	return c
}

func (p *Parser) StringSlice(key string, set OpSet) *StringSlice {
	c := new(StringSlice)
	p.Condition(key, set, c)
	return c
}

func (p *Parser) Int64(key string, set OpSet) *Int64 {
	c := new(Int64)
	p.Condition(key, set, c)
	return c
}

func (p *Parser) Int64Slice(key string, set OpSet) *Int64Slice {
	c := new(Int64Slice)
	p.Condition(key, set, c)
	return c
}

func (p *Parser) Timestamp(key string, set OpSet, loc *time.Location) *Timestamp {
	c := new(Timestamp)
	c.Location = loc
	p.Condition(key, set, c)
	return c
}

func (p *Parser) Bool(key string, set OpSet) *Bool {
	c := new(Bool)
	p.Condition(key, set, c)
	return c
}

func (p *Parser) Float64(key string, set OpSet) *Float64 {
	c := new(Float64)
	p.Condition(key, set, c)
	return c
}

type Condition interface {
	Set(key string, op Op, text []byte) error
}

type String struct {
	Key   string
	Op    Op
	Value string
}

func (c *String) Set(key string, op Op, text []byte) error {
	c.Key = key
	c.Op = op
	c.Value = string(text)
	return nil
}

type StringSlice struct {
	Key   string
	Op    Op
	Value []string
}

func (c *StringSlice) Set(key string, op Op, text []byte) error {
	values := bytes.Split(text, []byte(","))
	ss := make([]string, len(values))
	for i, v := range values {
		ss[i] = string(bytes.TrimSpace(v))
	}

	c.Key = key
	c.Op = op
	c.Value = ss
	return nil
}

type Int64 struct {
	Key   string
	Op    Op
	Value int64
}

func (c *Int64) Set(key string, op Op, text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 64)
	if err != nil {
		return err
	}

	c.Key = key
	c.Op = op
	c.Value = v
	return nil
}

type Int64Slice struct {
	Key   string
	Op    Op
	Value []int64
}

func (c *Int64Slice) Set(key string, op Op, text []byte) error {
	values := bytes.Split(text, []byte(","))
	is := make([]int64, len(values))
	for i, v := range values {
		iv, err := strconv.ParseInt(string(bytes.TrimSpace(v)), 10, 64)
		if err != nil {
			return err
		}
		is[i] = iv
	}

	c.Key = key
	c.Op = op
	c.Value = is
	return nil
}

type Timestamp struct {
	Key      string
	Op       Op
	Value    time.Time
	Location *time.Location
}

func (c *Timestamp) Set(key string, op Op, text []byte) error {
	loc := time.UTC
	if c.Location != nil {
		loc = c.Location
	}
	ts, err := time.ParseInLocation("2006-01-02 15:04:05", string(bytes.TrimSpace(text)), loc)
	if err != nil {
		return err
	}

	c.Key = key
	c.Op = op
	c.Value = ts
	return nil
}

type Bool struct {
	Key   string
	Op    Op
	Value bool
}

func (c *Bool) Set(key string, op Op, text []byte) error {
	ts, err := strconv.ParseBool(string(text))
	if err != nil {
		return err
	}

	c.Key = key
	c.Op = op
	c.Value = ts
	return nil
}

type Float64 struct {
	Key   string
	Op    Op
	Value float64
}

func (c *Float64) Set(key string, op Op, text []byte) error {
	v, err := strconv.ParseFloat(string(text), 64)
	if err != nil {
		return err
	}

	c.Key = key
	c.Op = op
	c.Value = v
	return nil
}
