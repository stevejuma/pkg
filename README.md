# pkg 

[![Github Workflow](https://img.shields.io/github/workflow/status/stevejuma/pkg/make%20test?style=for-the-badge)](https://github.com/stevejuma/pkg/actions/workflows/make_test.yml)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/stevejuma/pkg)
[![Codecov](https://img.shields.io/codecov/c/github/stevejuma/pkg?style=for-the-badge)](https://codecov.io/gh/stevejuma/pkg)

A home for various Go packages to be imported by other projects.

## Libraries 

 * [lucenequery](./lucenequery)
   * A Go parser for the lucene query language syntax
   * [lucenequery/sql](./lucenequery/sql)
      * A SQL code generator from the lucene query syntax
 * [fieldmask](./fieldmask)
   * A Go parser for partial response field mask queries loosely based on XPath