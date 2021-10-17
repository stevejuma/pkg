package sql

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"pkg/lucenequery"
	"regexp"
	"strings"
)


// PlaceHolder is the constant value used to indicate a variable substitution
const PlaceHolder = "?"

var operatorMappings = map[string]string{
	"eq":       "=",
	"gt":       ">",
	"gte":      ">=",
	"lt":       "<",
	"lte":      "<=",
	"neq":      "<>",
	"~":        "~",
	"~*":       "~*",
	"!~":       "!~",
	"!~*":      "!~*",
	"in":       "IN",
	"between":  "BETWEEN",
	"IMPLICIT": "OR",
	"AND":      "AND",
	"OR":       "OR",
	"NOT":      "NOT",
}

// Fragment a generated sql fragment with args
type Fragment struct {
	Column string
	Term   string
	Query  string
	Args   []interface{}
}

// InHandler is a handler for generating in values
type InHandler func(interface{}) interface{}

// ColumnHandler returns the true expression for the column
type ColumnHandler func(interface{}) (Fragment, error)

// SearchMode is the mode to apply searches in
type SearchMode int32

const (
	SearchModeAny SearchMode = 0
	SearchModeAll SearchMode = 1
)

// Enum value maps for SearchMode.
var (
	SearchModeName = map[int32]string{
		0: "ANY",
		1: "ALL",
	}
	SearchModeValue = map[string]int32{
		"ANY": 0,
		"ALL": 1,
	}
)

func (x SearchMode) Number() int32 {
	return int32(x)
}

func (x SearchMode) String() string {
	return SearchModeName[x.Number()]
}

func (x SearchMode) ValueOf(value string) SearchMode {
	return SearchMode(SearchModeValue[value])
}


// ToSQLOptions specifies properties for the ToSQL function
type ToSQLOptions struct {
	// Default field is the default column to use for filtering when not defined
	// If not provided, function will throw an error when a term without a name is encountered
	DefaultField string
	// SearchMode `ANY` increases the recall of queries by including more results,
	// and by default - will be interpreted as "OR NOT"
	// SearchMode `ALL` increases the precision of queries by including fewer results,
	// and by default - will be interpreted as "AND NOT"
	SearchMode SearchMode
	InHandler
	ColumnHandler
}

// Query is the generated query
type Query struct {
	Query   string
	Args    []interface{}
	Columns []string
}

var regexes = []struct {
	Pattern *regexp.Regexp
	Replace string
}{
	{Pattern: regexp.MustCompile(`^\s*(AND|OR)\s+([^()]+)(AND|OR)`), Replace: "$2$1"},
	{Pattern: regexp.MustCompile(`^\s*(AND|OR)\s*([^()]+)$`), Replace: "$2"},
	{Pattern: regexp.MustCompile(`("[^"]+").""`), Replace: "$1"},
}

// ToSQL returns the query as SQL string
func ToSQL(filter interface{}, opt *ToSQLOptions) (Query, error) {
	if opt.ColumnHandler == nil {
		opt.ColumnHandler = func(field interface{}) (Fragment, error) {
			switch f := field.(type) {
			case lucenequery.RangeQuery:
				return Fragment{Term: f.Term, Column: f.Term}, nil
			case lucenequery.TermQuery:
				return Fragment{Term: f.Term, Column: f.Term}, nil
			default:
				return Fragment{}, fmt.Errorf("unknonw type: %T", f)
			}
		}
	}
	query, err := renderSQL(filter, opt)
	if err != nil {
		return query, err
	}
	log.WithFields(log.Fields{
		"filter":  filter,
		"options": opt,
		"sql":     query.Query,
	}).Debug("SQL generated")
	query.Query = cleanExpr(query.Query)
	return query, err
}

func cleanExpr(expr string) string {
	for _, r := range regexes {
		expr = r.Pattern.ReplaceAllString(expr, r.Replace)
	}
	return strings.TrimSpace(expr)
}

func renderSQL(filter interface{}, opt *ToSQLOptions) (Query, error) {
	var query, cache = Query{Query: "", Args: []interface{}{}, Columns: []string{}}, map[string]string{}
	switch v := filter.(type) {
	case []interface{}:
		for _, r := range v {
			q, err := renderSQL(r, opt)
			if err != nil {
				return query, err
			}
			for _, t := range q.Columns {
				if _, ok := cache[t]; !ok {
					query.Columns = append(query.Columns, t)
				}
			}
			query.Args = append(query.Args, q.Args...)
			query.Query += q.Query
		}
		query.Query = cleanExpr(query.Query)
		return query, nil
	case string:
		dsl, err := lucenequery.Parse("ToSQL", []byte(v))
		if err != nil {
			return query, err
		}
		log.WithFields(log.Fields{
			"query":  v,
			"dsl":     dsl,
		}).Debug("Parsed Query")
		return renderSQL(dsl, opt)
	case lucenequery.BooleanExpression:
		size := len(v.Args)
		for i, r := range v.Args {
			q, err := renderSQL(r, opt)
			op := operatorMappings[v.Op]
			if v.Op == "IMPLICIT" && opt.SearchMode == SearchModeAll {
				op = "AND"
			}
			if op == "" {
				op = "OR"
			}
			if op == "NOT" {
				if opt.SearchMode == SearchModeAny {
					op = "OR NOT"
				} else {
					op = "AND NOT"
				}
			}

			if err != nil {
				return q, err
			}
			if i > 0 && i < size {
				if m, _ := regexp.MatchString(`^\s*(AND|OR|NOT)`, q.Query); !m {
					query.Query += fmt.Sprintf(" %s ", op)
				}
			}
			for _, t := range q.Columns {
				if _, ok := cache[t]; !ok {
					query.Columns = append(query.Columns, t)
				}
			}
			query.Query += q.Query
			query.Args = append(query.Args, q.Args...)
		}
		query.Query = fmt.Sprintf("(%s)", strings.TrimSpace(cleanExpr(query.Query)))
		return query, nil
	case lucenequery.TermQuery:
		fragment, err := opt.ColumnHandler(v)
		if err != nil {
			log.WithFields(log.Fields{
				"term": v.Term,
				"sql":  query.Query,
			}).Errorf("unknown column `%s`", v.Term)
			return query, fmt.Errorf("invalid column: `%s` error: %s", v.Term, err)
		}
		if fragment.Column != "" {
			query.Columns = append(query.Columns, fragment.Column)
		}
		if fragment.Query != "" {
			query.Query = fragment.Query
			query.Args = fragment.Args
			return query, nil
		}
		term := fragment.Term
		if term == "" {
			if opt != nil && opt.DefaultField != "" {
				term = opt.DefaultField
			} else {
				return query, fmt.Errorf("invalid term value `%v` provided for term without a name", v.Value)
			}
		}
		op := "="
		if v.Op != "" {
			if v, ok := operatorMappings[v.Op]; ok {
				op = v
			}
		}
		query.Query = fmt.Sprintf("%s %s %s", term, op, PlaceHolder)
		query.Args = []interface{}{v.Value}

		if v.Value == nil {
			op = "IS"
			query.Args = []interface{}{}
			query.Query = fmt.Sprintf("%s %s NULL", term, op)
			if v.Prefix == "-" {
				query.Query = fmt.Sprintf("%s %s NOT NULL", term, op)
			}
			return query, nil
		}

		if t, ok := v.Value.(lucenequery.WildCardQuery); ok {
			op = "LIKE"
			switch t.Kind() {
			case "prefix":
				query.Query = fmt.Sprintf("%s %s '%s%%'", term, op, PlaceHolder)
				query.Args = []interface{}{t.Prefix}
			case "suffix":
				query.Query = fmt.Sprintf("%s %s '%%%s'", term, op, PlaceHolder)
				query.Args = []interface{}{t.Suffix}
			case "between":
				query.Query = fmt.Sprintf("%s %s '%s%%%s'", term, op, PlaceHolder, PlaceHolder)
				query.Args = []interface{}{t.Prefix, t.Suffix}
			case "any":
				query.Query = fmt.Sprintf("%s %s '%%%s%%'", term, op, PlaceHolder)
				query.Args = []interface{}{t.Term}
			default:
				query.Query = fmt.Sprintf("%s IS NOT NULL", term)
				query.Args = []interface{}{}
			}
		}

		if op == "IN" {
			query.Query = fmt.Sprintf("%s %s (%s)", term, op, PlaceHolder)
			if opt.InHandler != nil {
				query.Args[0] = opt.InHandler(v.Value)
			}
			if t, ok := v.Value.([]interface{}); ok {
				if len(t) == 0 {
					query.Args = []interface{}{}
					query.Query = "1 = 0"
				}
			}
		}
		if v.Prefix == "+" {
			query.Query = fmt.Sprintf(" AND %s", query.Query)
		} else if v.Prefix == "-" {
			if opt.SearchMode == SearchModeAny {
				query.Query = fmt.Sprintf(" OR NOT %s", query.Query)
			} else {
				query.Query = fmt.Sprintf(" AND NOT %s", query.Query)
			}
		}
		return query, nil
	case lucenequery.RangeQuery:
		op, err := v.Kind()
		if err != nil {
			return query, fmt.Errorf("invalid column: `%s` error: %s", v.Term, err)
		}
		fragment, err := opt.ColumnHandler(v)
		if err != nil {
			log.WithFields(log.Fields{
				"term": v.Term,
				"sql":  fragment,
			}).Errorf("unknown column `%s`", v.Term)
			return query, fmt.Errorf("invalid column: `%s` error: %s", v.Term, err)
		}
		if fragment.Column != "" {
			query.Columns = append(query.Columns, fragment.Column)
		}
		if fragment.Query != "" {
			query.Query = fragment.Query
			query.Args = fragment.Args
			return query, nil
		}
		term := fragment.Term
		if term == "" {
			if opt != nil && opt.DefaultField != "" {
				term = opt.DefaultField
			} else {
				return query, fmt.Errorf("invalid range term value `%v` provided for term without a name", v)
			}
		}
		switch op {
		case "gt", "gte":
			query.Query = fmt.Sprintf("%s %s %s", term, operatorMappings[op], PlaceHolder)
			query.Args = []interface{}{v.Min}
			return query, nil
		case "lt", "lte":
			query.Query = fmt.Sprintf("%s %s %s", term, operatorMappings[op], PlaceHolder)
			query.Args = []interface{}{v.Max}
			return query, nil
		case "between":
			if v.Inclusive {
				query.Query = fmt.Sprintf("%s %s %s and %s", term, operatorMappings[op], PlaceHolder, PlaceHolder)
				query.Args = []interface{}{v.Min, v.Max}
				return query, nil
			}
			query.Query = fmt.Sprintf("%s > %s and %s < %s", term, PlaceHolder, term, PlaceHolder)
			query.Args = []interface{}{v.Min, v.Max}
			return query, nil
		default:
			return query, fmt.Errorf("unknown range type: %s", op)
		}
	default:
		return query, fmt.Errorf("unknown type: `%T`", v)
	}
}