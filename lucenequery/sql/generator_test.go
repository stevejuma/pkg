package sql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateSQL(t *testing.T) {
	cases := []struct {
		filter interface{}
		sql    string
		args   []interface{}
		opt    *ToSQLOptions
	}{
		{
			filter: `>= 5 <= 20`,
			sql:    `(id >= ? OR id <= ?)`,
			args:   []interface{}{5, 20},
			opt: &ToSQLOptions{
				DefaultField: "id",
			},
		},
		{
			filter: `user_id: +"google:001"`,
			sql:    `user_id = ?`,
			args:   []interface{}{"google:001"},
		},
		{
			filter: `user_id: -"2"`,
			sql:    `NOT user_id = ?`,
			args:   []interface{}{"2"},
		},
		{
			filter: `((age: > 18 age: <= 25) OR (age:[19,20])) NOT (age.teen:22 age.baby: [* TO 5])`,
			sql:    `(((age > ? OR age <= ?) OR age IN (?)) OR NOT (age.teen = ? OR age.baby <= ?))`,
			args:   []interface{}{18, 25, []interface{}{19, 20}, 22, 5},
		},
		{
			filter: `body:(+apple +mac)`,
			sql:    `(body = ? AND body = ?)`,
			args:   []interface{}{"apple", "mac"},
		},
		{
			filter: `body:(+apple -mac)`,
			sql:    `(body = ? AND NOT body = ?)`,
			args:   []interface{}{"apple", "mac"},
		},
		{
			filter: `age: null`, // available: true +disabled: false
			sql:    `age IS NULL`,
			args:   []interface{}{},
		},
		{
			filter: `age: -null`,
			sql:    `age IS NOT NULL`,
			args:   []interface{}{},
		},
		{
			filter: `name:(-null +"")`,
			sql:    `(name IS NOT NULL AND name = ?)`,
			args:   []interface{}{""},
		},
		{
			filter: `age: null available: true`,
			sql:    `(age IS NULL OR available = ?)`,
			args:   []interface{}{true},
		},
		{
			filter: `value: *`,
			sql:    `value IS NOT NULL`,
			args:   []interface{}{},
		},
		{
			filter: `value: term*`,
			sql:    `value LIKE '?%'`,
			args:   []interface{}{"term"},
		},
		{
			filter: `value: *term`,
			sql:    `value LIKE '%?'`,
			args:   []interface{}{"term"},
		},

		{
			filter: `value: te*m`,
			sql:    `value LIKE '?%?'`,
			args:   []interface{}{"te", "m"},
		},
		{
			filter: `value: *term*`,
			sql:    `value LIKE '%?%'`,
			args:   []interface{}{"term"},
		},
		{
			filter: `artists:(+"Miles Davis" -"John Coltrane" -"wayne")`,
			sql:    `(artists = ? AND (NOT artists = ? AND NOT artists = ?))`,
			args:   []interface{}{"Miles Davis", "John Coltrane", "wayne"},
			opt: &ToSQLOptions{
				DefaultField: "id",
				SearchMode:   SearchModeAll,
			},
		},
		{
			filter: `name: ~ "peter"`,
			sql:    `name ~ ?`,
			args:   []interface{}{"peter"},
		},
		{
			filter: `name: ~* "peter"`,
			sql:    `name ~* ?`,
			args:   []interface{}{"peter"},
		},
		{
			filter: `name: !~ "peter"`,
			sql:    `name !~ ?`,
			args:   []interface{}{"peter"},
		},
		{
			filter: `name: !~* "peter"`,
			sql:    `name !~* ?`,
			args:   []interface{}{"peter"},
		},
	}

	for _, dt := range cases {
		opt := &ToSQLOptions{DefaultField: ""}
		if dt.opt != nil {
			opt = dt.opt
		}
		query, err := ToSQL(dt.filter, opt)
		assert.NoError(t, err, dt)
		assert.Equal(t, dt.sql, query.Query, dt)
		assert.Equal(t, dt.args, query.Args, dt)
	}
}
