[![GoDoc](https://godoc.org/github.com/aryszka/tagstash?status.svg)](https://godoc.org/github.com/aryszka/tagstash)
[![Go Report Card](https://goreportcard.com/badge/github.com/aryszka/tagstash)](https://goreportcard.com/report/github.com/aryszka/tagstash)
[![Go Cover](https://gocover.io/_badge/github.com/aryszka/tagstash)](https://gocover.io/github.com/aryszka/tagstash)

# Tagstash

Tagstash is a library for tagging. It stores many-to-many associations between arbitrary values and tags, and
provides a best match lookup of values for multi-tag queries.

It is designed to decouple tagging from data. It typically stores URIs with one or more tag. The lookup
operation accepts multiple tags, and tries to find the stored value that is the closest match for the provided
set. Internally, it relies on a Postgresql based storage, extended with an in-memory cache. Both the persistent
storage and the cache accept custom implementations.

### Example:

```
stash := tagstash.New(tagstash.Options{
	CacheOptions: tagstash.CacheOptions{
		CacheSize: 1 << 12,
	},
})

stash.Set("https://www.example.org/page1.html", "foo", "bar", "baz")
stash.Set("https://www.example.org/page2.html", "foo", "qux", "quux")

if u, err := stash.Get("foo", "qux", "wah"); err != nil {
	fmt.Printf("error: %v", err)
} else {
	fmt.Printf("found: %s", u)
}
```

### Installation

```
go get github.com/aryszka/tagstash
```

### Documentation

Find the godoc documentation here:

[https://godoc.org/github.com/aryszka/tagstash](https://godoc.org/github.com/aryszka/tagstash)
