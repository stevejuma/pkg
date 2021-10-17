package fieldmask

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskExtract(t *testing.T) {
	cases := map[string]interface{}{
		"items,create_time":                         [][]string{{"items"}, {"create_time"}},
		"items ( id )":                              [][]string{{"items", "id"}},
		`labels(techaid.tech/uuid)`:                 [][]string{{"labels", "techaid.tech", "uuid"}},
		`labels("techaid.tech/uuid")`:               [][]string{{"labels", "techaid.tech/uuid"}},
		`labels/"techaid.tech/uuid"`:                [][]string{{"labels", "techaid.tech/uuid"}},
		`"labels/techaid.tech/uuid"`:                [][]string{{"labels/techaid.tech/uuid"}},
		"items(id)":                                 [][]string{{"items", "id"}},
		"context/facets/label":                      [][]string{{"context", "facets", "label"}},
		"context.facets.label,items(id)":            [][]string{{"context.facets.label"}, {"items", "id"}},
		"  links /* / href ":                        [][]string{{"links", "*", "href"}},
		"etag,items":                                [][]string{{"etag"}, {"items"}},
		"etag,items/title":                          [][]string{{"etag"}, {"items", "title"}},
		"items/name,items(title,author/uri),fields": [][]string{{"items", "name"}, {"items", "title"}, {"items", "author", "uri"}, {"fields"}},
		"items(title,author(uri(scheme/prefix)))":   [][]string{{"items", "title"}, {"items", "author", "uri", "scheme", "prefix"}},
		"context/facets/*(labels, pages)":           [][]string{{"context", "facets", "*", "labels"}, {"context", "facets", "*", "pages"}},
	}
	for q, expected := range cases {
		got, err := Masks(q)
		assert.NoError(t, err, q)
		assert.Equal(t, expected, got, q)
	}
}