package tagstash

import (
	"testing"

	"github.com/aryszka/keyval"
)

func newTestStash() *TagStash {
	return New(Options{
		Storage: &mockStorage{},
		CacheOptions: CacheOptions{
			CacheSize: 1 << 12,
		},
	})
}

func Test(t *testing.T) {
	t.Run("empty stash", func(t *testing.T) {
		stash := newTestStash()
		defer stash.Close()

		if v, err := stash.Get("foo"); err != nil || v != "" {
			t.Error("failed to query empty stash")
		}
	})

	t.Run("not found", func(t *testing.T) {
		stash := newTestStash()
		defer stash.Close()

		stash.Set("https://www.example.org", "foo")
		if v, err := stash.Get("bar"); err != nil || v != "" {
			t.Error("failed to query non-existing tag")
		}
	})

	t.Run("found", func(t *testing.T) {
		stash := newTestStash()
		defer stash.Close()

		stash.Set("https://www.example.org", "foo")
		if v, err := stash.Get("foo"); err != nil || v != "https://www.example.org" {
			t.Error("failed to query existing tag")
		}
	})

	t.Run("match with multiple tags", func(t *testing.T) {
		stash := newTestStash()
		defer stash.Close()

		stash.Set("https://www.example.org", "foo", "bar", "baz")
		if v, err := stash.Get("foo", "bar", "qux"); err != nil || v != "https://www.example.org" {
			t.Error("failed to query multiple tags")
		}
	})
}

func TestPriority(t *testing.T) {
	t.Run("match the one with more matching tags", func(t *testing.T) {
		stash := newTestStash()
		defer stash.Close()

		stash.Set("https://www.example.org/page1", "foo", "bar", "baz")
		stash.Set("https://www.example.org/page2", "foo", "qux", "quux")
		if v, err := stash.Get("foo", "bar", "qux", "quux"); err != nil || v != "https://www.example.org/page2" {
			t.Error("failed to query multiple tags")
		}
	})

	t.Run("match the one with closer tag order", func(t *testing.T) {
		stash := newTestStash()
		defer stash.Close()

		stash.Set("https://www.example.org/page1", "foo", "bar", "baz")
		stash.Set("https://www.example.org/page2", "bar", "foo", "qux")
		if v, err := stash.Get("foo", "bar", "baz", "qux", "quux"); err != nil || v != "https://www.example.org/page1" {
			t.Error("failed to query multiple tags")
		}
	})
}

func TestGetAllReturnsUniqueValues(t *testing.T) {
	stash := newTestStash()
	defer stash.Close()

	stash.Set("https://www.example.org/page1", "foo", "bar", "baz")
	stash.Set("https://www.example.org/page2", "foo", "bar", "qux")
	stash.Set("https://www.example.org/page3", "bar", "baz", "qux")

	v, err := stash.GetAll("foo")
	if err != nil {
		t.Error("failed to return multiple values")
	}

	if len(v) != 2 {
		t.Error("failed to return unique values")
	}
}

func TestDamagedCache(t *testing.T) {
	t.Run("full range damaged", func(t *testing.T) {
		stash := newTestStash()
		defer stash.Close()

		stash.Set("https://www.example.org", "foo", "bar", "baz")
		stash.cache.(*cache).forget.SetBytes("foo", []byte{'['}, forEver)
		if _, err := stash.Get("foo"); err == nil {
			t.Error("failed to detect damaged cache")
		}
	})

	t.Run("tag index damaged", func(t *testing.T) {
		stash := newTestStash()
		defer stash.Close()

		stash.Set("https://www.example.org", "foo", "bar", "baz")

		w, _ := stash.cache.(*cache).forget.Set("foo", forEver)
		defer w.Close()

		kvw := keyval.NewEntryWriter(w)
		kvw.WriteEntry(&keyval.Entry{
			Key: []string{"https://www.example.org"},
			Val: "a",
		})

		if _, err := stash.Get("foo"); err == nil {
			t.Error("failed to detect damaged cache")
		}
	})

	t.Run("value damaged", func(t *testing.T) {
		stash := newTestStash()
		defer stash.Close()

		stash.Set("https://www.example.org", "foo", "bar", "baz")

		w, _ := stash.cache.(*cache).forget.Set("foo", forEver)
		defer w.Close()

		kvw := keyval.NewEntryWriter(w)
		kvw.WriteEntry(&keyval.Entry{
			Key: []string{
				"https://www.example.org",
				"https://www.example.org/page1",
			},
			Val: "1",
		})

		if _, err := stash.Get("foo"); err == nil {
			t.Error("failed to detect damaged cache")
		}
	})

	t.Run("damaged entry on append", func(t *testing.T) {
		stash := newTestStash()
		defer stash.Close()

		stash.Set("https://www.example.org/page1", "foo", "bar", "baz")
		stash.cache.(*cache).forget.SetBytes("foo", []byte{'['}, forEver)

		if err := stash.Set("https://www.example.org/page2", "foo"); err == nil {
			t.Error("failed to detect damaged cache")
		}
	})
}

func TestOversize(t *testing.T) {
	t.Run("tag too large", func(t *testing.T) {
		stash := New(Options{
			Storage: &mockStorage{},
			CacheOptions: CacheOptions{
				CacheSize:        1 << 8,
				ExpectedItemSize: 1 << 6,
			},
		})
		defer stash.Close()

		large := make([]byte, 512)
		for i := range large {
			large[i] = 42
		}

		if err := stash.Set("123456789", string(large)); err == nil {
			t.Error("failed to fail")
		}
	})

	t.Run("value too large", func(t *testing.T) {
		stash := New(Options{
			Storage: &mockStorage{},
			CacheOptions: CacheOptions{
				CacheSize:        1 << 8,
				ExpectedItemSize: 1 << 6,
			},
		})
		defer stash.Close()

		large := make([]byte, 512)
		for i := range large {
			large[i] = 42
		}

		if err := stash.Set(string(large), "123456"); err == nil {
			t.Error("failed to fail")
		}
	})
}

// func TestStorageFails(t *testing.T) {
// 	t.Run("get", func(t *testing.T) {
// 		stash := newTestStash()
// 		defer stash.Close()
//
// 		if err := stash.Set("https://www.example.org", "foo"); err != nil {
// 			t.Error("failed to set initial item")
// 			return
// 		}
//
// 		// stash.cache.(*cache)
//
// 		stash.storage.(*mockStorage).failNext = true
// 		if _, err := stash.Get("foo"); err == nil {
// 			t.Error("failed to fail", v, err)
// 		}
// 	})
//
// 	t.Run("set", func(t *testing.T) {
// 		stash := newTestStash()
// 		defer stash.Close()
//
// 		stash.storage.(*mockStorage).failNext = true
// 		if err := stash.Set("https://www.example.org", "foo"); err == nil {
// 			t.Error("failed to fail")
// 		}
//
// 		if v, err := stash.Get("foo"); err != nil || v != "" {
// 			t.Error("failed to fail", v, err)
// 		}
// 	})
// }

func TestReorder(t *testing.T) {
	stash := newTestStash()
	defer stash.Close()

	stash.Set("https://www.example.org/page1", "foo", "bar", "baz")
	stash.Set("https://www.example.org/page2", "baz", "bar", "foo")

	if v, err := stash.Get("foo", "baz", "bar"); err != nil || v != "https://www.example.org/page1" {
		t.Error("failed to get the right initial value", v, err)
	}

	stash.Set("https://www.example.org/page2", "foo", "baz", "bar")

	if v, err := stash.Get("foo", "baz", "bar"); err != nil || v != "https://www.example.org/page2" {
		t.Error("failed to get the right reordered value", v, err)
	}
}

func TestRemoveTag(t *testing.T) {
	stash := newTestStash()
	defer stash.Close()

	stash.Set("https://www.example.org", "foo", "bar", "baz")
	if err := stash.Remove("https://www.example.org", "foo"); err != nil {
		t.Error("failed to remove tag", err)
	}

	if v, err := stash.Get("foo"); err != nil || v != "" {
		t.Error("failed to remove tag", v, err)
	}
}

func TestClearTag(t *testing.T) {
	stash := newTestStash()
	defer stash.Close()

	stash.Set("https://www.example.org/page1", "foo", "bar")
	stash.Set("https://www.example.org/page2", "foo", "baz")

	if err := stash.Delete("foo"); err != nil {
		t.Error("failed to delete tag", err)
	}

	if v, err := stash.Get("foo"); err != nil || v != "" {
		t.Error("failed to delete tag", err)
	}

	if v, err := stash.Get("bar"); err != nil || v != "https://www.example.org/page1" {
		t.Error("failed to keep tag: bar", err, v)
	}

	if v, err := stash.Get("baz"); err != nil || v != "https://www.example.org/page2" {
		t.Error("failed to keep tag: baz", err, v)
	}
}
