package lucenequery

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/google/go-cmp/cmp"
)

type TestCase struct {
	expected interface{}
	queries  []string
}

func toJSON(t *testing.T, v interface{}) string {
	jsonData, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Expected to marshal %v without error, got: %v", v, err)
	}
	return string(jsonData)
}

func typeOf(v interface{}) reflect.Type{
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		return typeOf(val.Elem().Interface())
	}
	return reflect.TypeOf(v)
}

func executeTestCases(t *testing.T, cases []TestCase) {
	for i, test := range cases {
		for j, q := range test.queries {
			got, err := Parse("TestScalarValues", []byte(q))
			if err != nil {
				t.Fatalf("Expected to parse %s without error, got: %v", q, err)
			}
			expectedType, gotType := typeOf(test.expected), typeOf(got)
			expectedValue, gotValue := toJSON(t, test.expected), toJSON(t, got)
			if !cmp.Equal(expectedValue, gotValue) || (expectedType != gotType) {
				t.Fatalf("[%v,%v]Expected lucenequery {%s}`%s` to equal {%s}\n\t%s, \n\t\t got \n\t%s\n\t\t diff \n\t%s",
					i, j, expectedType, q, gotType, expectedValue, gotValue, diff.CharacterDiff(expectedValue, gotValue))
			}
		}
	}
}

func TestTermQueries(t *testing.T) {
	executeTestCases(t, []TestCase{
		{
			queries:  []string{`name: peter`, `name: "peter"`},
			expected: &TermQuery{Term: "name", Value: "peter", Op: ""},
		},
		{
			queries:  []string{`labels.tech.tech/volunteers/type: "peter"`, `labels.tech.tech/volunteers/type: peter`},
			expected: &TermQuery{Term: "labels.tech.tech/volunteers/type", Value: "peter", Op: ""},
		},
		{
			queries:  []string{`name: eq "peter"`},
			expected: &TermQuery{Term: "name", Value: "peter", Op: "eq"},
		},
		{
			queries:  []string{`age: null`},
			expected: &TermQuery{Term: "age", Value: nil, Op: ""},
		},
		{
			queries:  []string{`available: false`},
			expected: &TermQuery{Term: "available", Value: false, Op: ""},
		},
		{
			queries:  []string{`available: true`},
			expected: &TermQuery{Term: "available", Value: true, Op: ""},
		},
		{
			queries:  []string{`age: 23`},
			expected: &TermQuery{Term: "age", Value: 23, Op: ""},
		},
		{
			queries:  []string{`metric: -23`},
			expected: &TermQuery{Term: "metric", Value: -23, Op: ""},
		},
		{
			queries:  []string{`age: 23.5`},
			expected: &TermQuery{Term: "age", Value: 23.5, Op: ""},
		},
		{
			queries:  []string{`metric: -123.456`},
			expected: &TermQuery{Term: "metric", Value: -123.456, Op: ""},
		},
		{
			queries:  []string{`quote: "a walk in the \"park\""`},
			expected: &TermQuery{Term: "quote", Value: `a walk in the "park"`, Op: ""},
		},
		{
			queries:  []string{`array: [1,-2.5,3.14,-12,"arrays"]`},
			expected: &TermQuery{Term: "array", Value: []interface{}{1, -2.5, 3.14, -12, "arrays"}, Op: "in"},
		},
	})
}

func TestRangeQueries(t *testing.T) {
	executeTestCases(t, []TestCase{
		{
			queries:  []string{`age: [18 TO 25]`},
			expected: RangeQuery{Min: 18, Max: 25, Term: "age", Inclusive: true},
		},
		{
			queries:  []string{`metric: [-18.54 TO 5.5]`},
			expected: RangeQuery{Min: -18.54, Max: 5.5, Term: "metric", Inclusive: true},
		},
		{
			queries:  []string{`age: {18 TO 25}`},
			expected: RangeQuery{Min: 18, Max: 25, Term: "age", Inclusive: false},
		},
		{
			queries:  []string{`metric: {-18.54 TO 5.5}`},
			expected: RangeQuery{Min: -18.54, Max: 5.5, Term: "metric", Inclusive: false},
		},
		{
			queries:  []string{`metric: ["2020-01-01" TO "2020-03-31"]`},
			expected: RangeQuery{Min: "2020-01-01", Max: "2020-03-31", Term: "metric", Inclusive: true},
		},
		{
			queries:  []string{`metric: {"2020-01-01" TO "2020-03-31"}`},
			expected: RangeQuery{Min: "2020-01-01", Max: "2020-03-31", Term: "metric", Inclusive: false},
		},
		{
			queries:  []string{`metric: [5 TO *]`, `metric: >= 5`, `metric: gte 5`},
			expected: RangeQuery{Min: 5, Max: "*", Term: "metric", Inclusive: true},
		},
		{
			queries:  []string{`metric: {5 TO *}`, `metric: > 5`, `metric: gt 5`},
			expected: RangeQuery{Min: 5, Max: "*", Term: "metric", Inclusive: false},
		},
		{
			queries:  []string{`metric: [* TO 3.14]`, `metric: <= 3.14`, `metric: lte 3.14`},
			expected: RangeQuery{Min: "*", Max: 3.14, Term: "metric", Inclusive: true},
		},
		{
			queries:  []string{`metric: {* TO 3.14}`, `metric: < 3.14`, `metric: lt 3.14`},
			expected: RangeQuery{Min: "*", Max: 3.14, Term: "metric", Inclusive: false},
		},
		{
			queries:  []string{`metric: {* TO *}`},
			expected: RangeQuery{Min: "*", Max: "*", Term: "metric", Inclusive: false},
		},
	})
}

func TestBooleanQueries(t *testing.T) {
	executeTestCases(t, []TestCase{
		{
			queries: []string{` OR `},
			expected: BooleanExpression{
				Op: "OR",
			},
		},
		{
			queries: []string{`OR AND`},
			expected: BooleanExpression{
				Op: "AND",
			},
		},
		{
			queries: []string{`OR AND foo`},
			expected: TermQuery{
				Value: "foo",
			},
		},
		{
			queries: []string{`"jakarta apache" jakarta`},
			expected: BooleanExpression{
				Op: "IMPLICIT",
				Args: []interface{}{
					TermQuery{Op: "", Value: "jakarta apache"},
					TermQuery{Op: "", Value: "jakarta"},
				},
			},
		},
		{
			queries: []string{`"jakarta apache" OR jakarta`, `"jakarta apache" || jakarta`},
			expected: BooleanExpression{
				Op: "OR",
				Args: []interface{}{
					TermQuery{Op: "", Value: "jakarta apache"},
					TermQuery{Op: "", Value: "jakarta"},
				},
			},
		},
		{
			queries: []string{`+jakarta lucene`},
			expected: BooleanExpression{
				Op: "IMPLICIT",
				Args: []interface{}{
					TermQuery{Op: "", Prefix: "+", Value: "jakarta"},
					TermQuery{Op: "", Value: "lucene"},
				},
			},
		},
		{
			queries: []string{`"jakarta apache" AND "Apache Lucene"`, `"jakarta apache" && "Apache Lucene"`},
			expected: BooleanExpression{
				Op: "AND",
				Args: []interface{}{
					TermQuery{Op: "", Value: "jakarta apache"},
					TermQuery{Op: "", Value: "Apache Lucene"},
				},
			},
		},
		{
			queries: []string{`"jakarta apache" NOT "Apache Lucene"`},
			expected: BooleanExpression{
				Op: "NOT",
				Args: []interface{}{
					TermQuery{Op: "", Value: "jakarta apache"},
					TermQuery{Op: "", Value: "Apache Lucene"},
				},
			},
		},
		{
			queries:  []string{`NOT "Apache Lucene"`},
			expected: TermQuery{Op: "", Value: "Apache Lucene"},
		},
		{
			queries: []string{`title:(+return +"pink panther")`},
			expected: BooleanExpression{
				Op: "IMPLICIT",
				Args: []interface{}{
					TermQuery{Op: "", Term: "title", Prefix: "+", Value: "return"},
					TermQuery{Op: "", Term: "title", Prefix: "+", Value: "pink panther"},
				},
			},
		},
		{
			queries: []string{`(jakarta OR apache) AND website`},
			expected: BooleanExpression{
				Op: "AND",
				Args: []interface{}{
					BooleanExpression{
						Op: "OR",
						Args: []interface{}{
							TermQuery{Op: "", Value: "jakarta"},
							TermQuery{Op: "", Value: "apache"},
						},
					},
					TermQuery{Op: "", Value: "website"},
				},
			},
		},
	})
}

func TestWildCardQueries(t *testing.T) {
	executeTestCases(t, []TestCase{
		{
			queries: []string{`*`},
			expected: TermQuery{
				Value: WildCardQuery{},
			},
		},
		{
			queries: []string{`title: jakat*`},
			expected: TermQuery{
				Term: "title",
				Value: WildCardQuery{
					Prefix: "jakat",
				},
			},
		},
		{
			queries: []string{`test*`},
			expected: TermQuery{
				Value: WildCardQuery{
					Prefix: "test",
				},
			},
		},
		{
			queries: []string{`tes*t`},
			expected: TermQuery{
				Value: WildCardQuery{
					Prefix: "tes",
					Suffix: "t",
				},
			},
		},
		{
			queries: []string{`tes?t`},
			expected: TermQuery{
				Value: "tes?t",
			},
		},
	})
}