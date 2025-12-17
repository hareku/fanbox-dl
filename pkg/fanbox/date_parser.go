package fanbox

import (
	"fmt"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

type Op int

const (
	OpLT Op = iota
	OpLTE
	OpGT
	OpGTE
	OpEQ
)

type Clause struct {
	Op   Op
	Date time.Time
}

type DatePredicate struct {
	Clauses []Clause
}

func parseDate(s string) (time.Time, error) {
	return time.Parse(dateLayout, strings.TrimSpace(s))
}

func parseOp(s string) (Op, error) {
	switch s {
	case "<":
		return OpLT, nil
	case "<=":
		return OpLTE, nil
	case ">":
		return OpGT, nil
	case ">=":
		return OpGTE, nil
	case "==":
		return OpEQ, nil
	default:
		return 0, fmt.Errorf("invalid operator %q", s)
	}
}

func invert(op Op) Op {
	switch op {
	case OpLT:
		return OpGT
	case OpLTE:
		return OpGTE
	case OpGT:
		return OpLT
	case OpGTE:
		return OpLTE
	default:
		return op
	}
}

func apply(op Op, a, b time.Time) bool {
	switch op {
	case OpLT:
		return a.Before(b)
	case OpLTE:
		return a.Before(b) || a.Equal(b)
	case OpGT:
		return a.After(b)
	case OpGTE:
		return a.After(b) || a.Equal(b)
	case OpEQ:
		return a.Equal(b)
	default:
		return false
	}
}

func parseDateOnly(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()), nil
	}

	if t, err := time.Parse(dateLayout, s); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid date string: %s", s)
}

func (p *DatePredicate) Matches(t time.Time) bool {
	if p == nil {
		return true
	}

	for _, c := range p.Clauses {
		if !apply(c.Op, t, c.Date) {
			return false
		}
	}
	return true
}

func parseLeftSide(s string) (Clause, error) {
	ops := []string{"<=", ">=", "<", ">"}

	for _, op := range ops {
		if strings.HasSuffix(s, op) {
			d, err := parseDate(s[:len(s)-len(op)])
			if err != nil {
				return Clause{}, err
			}
			parsedOp, _ := parseOp(op)
			return Clause{
				Op:   invert(parsedOp),
				Date: d,
			}, nil
		}
	}

	return Clause{}, fmt.Errorf("invalid left side %q", s)
}

func parseRightSide(s string) (Clause, error) {
	ops := []string{"<=", ">=", "<", ">"}

	for _, op := range ops {
		if strings.HasPrefix(s, op) {
			d, err := parseDate(s[len(op):])
			if err != nil {
				return Clause{}, err
			}
			parsedOp, _ := parseOp(op)
			return Clause{
				Op:   parsedOp,
				Date: d,
			}, nil
		}
	}

	return Clause{}, fmt.Errorf("invalid right side %q", s)
}

func parseChained(s string) (*DatePredicate, error) {
	s = strings.ReplaceAll(s, " ", "")
	parts := strings.Split(s, "x")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid chained expression %q", s)
	}

	var clauses []Clause

	if parts[0] != "" {
		c, err := parseLeftSide(parts[0])
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, c)
	}

	if parts[1] != "" {
		c, err := parseRightSide(parts[1])
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, c)
	}

	if len(clauses) == 0 {
		return nil, fmt.Errorf("empty chained expression %q", s)
	}

	return &DatePredicate{Clauses: clauses}, nil
}

func parseSingle(s string) (*DatePredicate, error) {
	ops := []string{"<=", ">=", "==", "<", ">"}

	for _, op := range ops {
		if strings.HasPrefix(s, op) {
			d, err := parseDate(s[len(op):])
			if err != nil {
				return nil, err
			}
			parsedOp, _ := parseOp(op)
			return &DatePredicate{
				Clauses: []Clause{{Op: parsedOp, Date: d}},
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid date expression %q", s)
}

func ParseDatePredicate(input string) (*DatePredicate, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	if !strings.ContainsAny(input, "<>=") {
		d, err := parseDate(input)
		if err != nil {
			return nil, err
		}
		return &DatePredicate{
			Clauses: []Clause{{Op: OpGTE, Date: d}},
		}, nil
	}

	if strings.Contains(input, "x") {
		return parseChained(input)
	}

	return parseSingle(input)
}
