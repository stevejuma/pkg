# Lucene Query SQL Generator 

A simple library that takes a predicate in the lucene query syntax format
and converts it to a sql query with placeholders replaced. 

```go
package querytest

query, err := ToSQL("body:(+apple +mac)", &ToSQLOptions{
    DefaultField: "id",
})
if err != nil {
    log.Fatalf("Error while parsing query: %v", err)
}
query == {
    filter: `body:(+apple +mac)`,
    sql:    `(body = ? AND body = ?)`,
    args:   []interface{}{"apple", "mac"},
}
```