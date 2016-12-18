package tagstash_test

import (
	"fmt"

	"github.com/aryszka/tagstash"
)

func Example() {
	stash := tagstash.New(tagstash.Options{
		CacheOptions: tagstash.CacheOptions{
			CacheSize: 1 << 12,
		},
	})

	stash.Set("https://www.example.org/page1.html", "foo", "bar", "baz")
	stash.Set("https://www.example.org/page2.html", "foo", "qux", "quux")

	if u, err := stash.Get("qux", "foo", "wah"); err != nil {
		fmt.Printf("error: %v", err)
	} else {
		fmt.Printf("found: %s", u)
	}

	// Output:
	// found: https://www.example.org/page2.html
}
