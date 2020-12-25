package query

import (
	"bytes"
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
	delimiter  []byte
	conditions []parseCondition
}

func NewParser(delimiter string) *Parser {
	return &Parser{
		[]byte(delimiter),
		nil,
	}
}

func (p *Parser) Parse(query string) error {
	data := []byte(query)
	exprs := bytes.Split(data, p.delimiter)

	for _, expr := range exprs {
		err := p.parseCondition(expr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) parseCondition(c []byte) error {
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

func (p *Parser) String(key string, set OpSet) *StringCondition {
	c := new(StringCondition)
	p.condition(key, set, c)
	return c
}

func (p *Parser) StringSlice(key string, set OpSet) *StringSliceCondition {
	c := new(StringSliceCondition)
	p.condition(key, set, c)
	return c
}

type Condition interface {
	Set(key string, op Op, text []byte) error
}

type StringCondition struct {
	Key   string
	Op    Op
	Value string
}

func (c *StringCondition) Set(key string, op Op, text []byte) error {
	c.Key = key
	c.Op = op
	c.Value = string(text)
	return nil
}

type StringSliceCondition struct {
	Key   string
	Op    Op
	Value []string
}

func (c *StringSliceCondition) Set(key string, op Op, text []byte) error {
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

func (p *Parser) condition(key string, set OpSet, c Condition) {
	p.conditions = append(p.conditions, parseCondition{[]byte(key), set, c})
}

type parseCondition struct {
	key []byte
	set OpSet
	c   Condition
}
