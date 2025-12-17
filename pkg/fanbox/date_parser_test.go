package fanbox

import (
	"testing"
	"time"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse(dateLayout, s)
	if err != nil {
		t.Fatalf("failed to parse date %q: %v", s, err)
	}
	return d
}

func TestParseDatePredicate_Empty(t *testing.T) {
	p, err := ParseDatePredicate("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != nil {
		t.Fatalf("expected nil predicate, got %#v", p)
	}
}

func TestParseDatePredicate_EqualityShorthand(t *testing.T) {
	p, err := ParseDatePredicate("2024-01-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.Clauses) != 1 {
		t.Fatalf("expected 1 clause, got %d", len(p.Clauses))
	}

	c := p.Clauses[0]
	if c.Op != OpGTE {
		t.Fatalf("expected OpEQ, got %v", c.Op)
	}
	if !c.Date.Equal(mustDate(t, "2024-01-01")) {
		t.Fatalf("unexpected date: %v", c.Date)
	}
}

func TestParseDatePredicate_SingleOperators(t *testing.T) {
	tests := []struct {
		input string
		op    Op
		date  string
	}{
		{"<2024-01-01", OpLT, "2024-01-01"},
		{"<=2024-01-01", OpLTE, "2024-01-01"},
		{">2024-01-01", OpGT, "2024-01-01"},
		{">=2024-01-01", OpGTE, "2024-01-01"},
		{"==2024-01-01", OpEQ, "2024-01-01"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p, err := ParseDatePredicate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(p.Clauses) != 1 {
				t.Fatalf("expected 1 clause, got %d", len(p.Clauses))
			}

			c := p.Clauses[0]
			if c.Op != tt.op {
				t.Fatalf("expected op %v, got %v", tt.op, c.Op)
			}
			if !c.Date.Equal(mustDate(t, tt.date)) {
				t.Fatalf("unexpected date: %v", c.Date)
			}
		})
	}
}

func TestParseDatePredicate_Chained(t *testing.T) {
	tests := []struct {
		input   string
		clauses []Clause
	}{
		{
			input: "2024-01-01<x<=2024-12-31",
			clauses: []Clause{
				{Op: OpGT, Date: mustDate(t, "2024-01-01")},
				{Op: OpLTE, Date: mustDate(t, "2024-12-31")},
			},
		},

		{
			input: "2024-01-01<x",
			clauses: []Clause{
				{Op: OpGT, Date: mustDate(t, "2024-01-01")},
			},
		},
		{
			input: "x<=2024-01-01",
			clauses: []Clause{
				{Op: OpLTE, Date: mustDate(t, "2024-01-01")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p, err := ParseDatePredicate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(p.Clauses) != len(tt.clauses) {
				t.Fatalf("expected %d clauses, got %d", len(tt.clauses), len(p.Clauses))
			}

			for i := range tt.clauses {
				if p.Clauses[i].Op != tt.clauses[i].Op {
					t.Fatalf("clause %d: expected op %v, got %v",
						i, tt.clauses[i].Op, p.Clauses[i].Op)
				}
				if !p.Clauses[i].Date.Equal(tt.clauses[i].Date) {
					t.Fatalf("clause %d: expected date %v, got %v",
						i, tt.clauses[i].Date, p.Clauses[i].Date)
				}
			}
		})
	}
}

func TestPredicate_Matches(t *testing.T) {
	p, err := ParseDatePredicate("2024-01-01<x<=2024-01-10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		date  string
		match bool
	}{
		{"2024-01-01", false},
		{"2024-01-02", true},
		{"2024-01-10", true},
		{"2024-01-11", false},
	}

	for _, tt := range tests {
		t.Run(tt.date, func(t *testing.T) {
			d := mustDate(t, tt.date)
			if got := p.Matches(d); got != tt.match {
				t.Fatalf("expected match=%v, got %v", tt.match, got)
			}
		})
	}
}

func TestParseDatePredicate_Invalid(t *testing.T) {
	tests := []string{
		"foo",
		"><2024-01-01",
		"2024-01-01<>2024-01-02",
		"2024-99-99",
		"2024-01-01x2024-01-02",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if _, err := ParseDatePredicate(input); err == nil {
				t.Fatalf("expected error for input %q", input)
			}
		})
	}
}
