# Field Mask 

Field mask is a library for parsing partial response request queries. 
This can be used with API's to improve performance and return only 
requested data.

## Fields mask parameter syntax 

The format of the `fields` request parameter value is loosely 
based on XPath syntax. The supported syntax is summarized below, 
and additional examples are provided in the following section.

* Use a comma-separated list to select multiple fields.
* Use `a/b` to select a field `b` that is nested within field `a`; 
* use `a/b/c` to select a field `c` nested within `b`
* Use a sub-selector to request a set of specific sub-fields of arrays 
  or objects by placing expressions in parentheses `"( )"`.

For example: `fields=items(id,author/email)` returns only the item ID 
and author's email for each element in the items array. You can also 
specify a single sub-field, where `fields=items(id) `is equivalent to `fields=items/id`.

* Use wildcards in field selections, if needed.
  For example: `fields=items/pagemap/*` selects all objects in a pagemap. 
* You can also omit the wildcard if it's at the end of the selector. 
  The above is similar to `fields=items/pagemap`

**Identify the fields you want returned, or make field selections.**

* `items`
    * Returns all elements in the items array, including all fields in  
      each element, but no other fields.

* `etag,items`
    * Returns both the **etag** field and all elements in the items array.

* `items/title`
    * Returns only the **title** field for all elements in the items array  
      Whenever a nested field is returned, the response includes the enclosing  
      parent objects. The parent fields do not include any other child fields 
      unless they are also selected explicitly

* `context/facets/label`
    * Returns only the **label** field for all members of the **facets** array,
      which is itself nested under the **context** object.

* `items/pagemap/*/title`
    * For each element in the items array, returns only the **title** field  (if present) 
      of all objects that are children of **pagemap**.

* `title`
    * Returns the `title` field of the requested resource.

* `author/uri`
    * Returns the `uri` sub-field of the `author` object in the requested resource.

* `links/*/href`
    * Returns the `href` field of all objects that are children of `links`.

* `items(title,author/uri)`
    * Returns only the values of the `title` and author's `uri` for each element in the items array.