package tagstash

import (
	"errors"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/aryszka/forget"
	"github.com/aryszka/keyval"
)

const forEver = time.Duration((^uint64(0)) >> 1)

type cache struct {
	forget *forget.Cache
	mx     *sync.Mutex
}

var (
	// ErrDamagedCacheData is returned when the cache detects damaged data.
	ErrDamagedCacheData = errors.New("damaged cache data")

	// ErrFailedToCacheEntry is returned when caching an entry failed, e.g. due to oversize.
	ErrFailedToCacheEntry = errors.New("failed to cache entry")
)

func newCache(o CacheOptions) *cache {
	if o.ExpectedItemSize < 64 {
		o.ExpectedItemSize = 64
	}

	return &cache{
		forget: forget.New(forget.Options{
			CacheSize: o.CacheSize,
			ChunkSize: o.ExpectedItemSize,
		}),
		mx: &sync.Mutex{},
	}
}

func readAll(r io.Reader, tag string) ([]*Entry, error) {
	var entries []*Entry
	kvr := keyval.NewEntryReader(r)
	for {
		e, err := kvr.ReadEntry()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		tagIndex, err := strconv.Atoi(e.Val)
		if err != nil {
			return nil, err
		}

		if len(e.Key) != 1 {
			return nil, ErrDamagedCacheData
		}

		entries = append(entries, &Entry{
			Value:    e.Key[0],
			Tag:      tag,
			TagIndex: tagIndex,
		})
	}

	return entries, nil
}

func writeAll(w io.Writer, e []*Entry) error {
	kvw := keyval.NewEntryWriter(w)
	for _, ei := range e {
		err := kvw.WriteEntry(&keyval.Entry{
			Key: []string{ei.Value},
			Val: strconv.Itoa(ei.TagIndex),
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *cache) withTagEntries(tag string, op func([]*Entry) []*Entry) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	var entries []*Entry
	if r, ok := c.forget.Get(tag); ok {
		defer r.Close()

		var err error
		entries, err = readAll(r, tag)
		if err != nil {
			return err
		}
	}

	entries = op(entries)

	w, ok := c.forget.Set(tag, forEver)
	if !ok {
		return ErrFailedToCacheEntry
	}

	defer w.Close()
	return writeAll(w, entries)
}

func (c *cache) Get(tags []string) ([]*Entry, error) {
	var entries []*Entry
	for _, t := range tags {
		r, ok := c.forget.Get(t)
		if !ok {
			continue
		}

		defer r.Close()
		tagEntries, err := readAll(r, t)
		if err != nil {
			return nil, err
		}

		entries = append(entries, tagEntries...)
	}

	return entries, nil
}

func (c *cache) Set(e *Entry) error {
	return c.withTagEntries(e.Tag, func(entries []*Entry) []*Entry {
		var exists bool
		for _, ei := range entries {
			if ei.Value == e.Value {
				ei.TagIndex = e.TagIndex
				exists = true
				break
			}
		}

		if !exists {
			entries = append(entries, e)
		}

		return entries
	})
}

func (c *cache) Remove(e *Entry) error {
	return c.withTagEntries(e.Tag, func(entries []*Entry) []*Entry {
		for i, ei := range entries {
			if ei.Value == e.Value {
				last := len(entries) - 1
				entries[last], entries = nil, append(entries[:i], entries[i+1:]...)
				break
			}
		}

		return entries
	})
}

func (c *cache) Delete(tag string) error {
	c.forget.Delete(tag)
	return nil
}

func (c *cache) Close() {
	c.forget.Close()
}
