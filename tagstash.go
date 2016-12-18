package tagstash

import "sort"

// Entry represents a value-tag associaction.
type Entry struct {

	// Value that a tag belongs to.
	Value string

	// Tag associated with a value.
	Tag string

	// TagIndex marks how strong strong a tag describes a value.
	TagIndex int

	requestTagCount, requestTagIndex int
}

// Storage implementations store value-tag associations.
type Storage interface {

	// Get returns all entries whose tag is listed in the arguments.
	Get([]string) ([]*Entry, error)

	// Set stores a value-tag association. Implementations must make sure that the value-tag combinations
	// are unique.
	Set(*Entry) error

	// Remove deletes a single value-tag association.
	Remove(*Entry) error

	// Delete deletes all associations with the provided tag.
	Delete(string) error
}

// StorageOptions are used by the default storage implementation.
type StorageOptions struct {
}

// CacheOptions are used by the default cache implementation.
type CacheOptions struct {

	// CacheSize defines the maximum memory usage of the cache. Defaults to 1G.
	CacheSize int

	// ExpectedValueSize provides a hint for the cache about the expected median size of the stored values.
	//
	// This option exists only for optimization, there is no good rule of thumb. Too high values will result
	// in worse memory utilization, while too low values may affect the individual lookup performance.
	// Generally, it is better to err for the smaller values.
	ExpectedValueSize int
}

// Options are used to initialization tagstash.
type Options struct {

	// Custom storage implementation. By default, a builtin storage is used.
	Storage Storage

	// Custom cache implementation. By default, a builtin cache is used.
	Cache Storage

	// CacheOptions define options for the default persistent storage implementation when not replaced by a custom
	// storage.
	StorageOptions StorageOptions

	// CacheOptions define options for the default cache implementation when not replaced by a custom
	// cache.
	CacheOptions CacheOptions
}

type entrySort struct {
	tags    []string
	entries []*Entry
}

// TagStash is used to store tags associated with values and return the best matching value for a set of query
// tags.
type TagStash struct {
	cache, storage Storage
}

func (e *Entry) matchValue() int {
	v := e.requestTagIndex - e.TagIndex
	if v < 0 {
		return 0 - v
	}

	return v
}

func (s entrySort) Len() int      { return len(s.entries) }
func (s entrySort) Swap(i, j int) { s.entries[i], s.entries[j] = s.entries[j], s.entries[i] }

func (s entrySort) Less(i, j int) bool {
	left, right := s.entries[i], s.entries[j]

	if left.requestTagCount == right.requestTagCount {
		return left.matchValue() < right.matchValue()
	}

	return left.requestTagCount > right.requestTagCount
}

// New creates and initializes a tagstash instance.
func New(o Options) *TagStash {
	if o.Storage == nil {
		o.Storage = newStorage(o.StorageOptions)
	}

	if o.Cache == nil {
		o.Cache = newCache(o.CacheOptions)
	}

	return &TagStash{
		storage: o.Storage,
		cache:   o.Cache,
	}
}

func setRequestIndex(tags []string, e []*Entry) (notFound []string) {
	for i, t := range tags {
		var found bool
		for _, ei := range e {
			if ei.Tag == t {
				ei.requestTagIndex = i
				found = true
			}
		}

		if !found {
			notFound = append(notFound, t)
		}
	}

	return notFound
}

func uniqueValues(e []*Entry) []*Entry {
	m := make(map[string]*Entry)
	u := make([]*Entry, 0, len(e))
	for _, ei := range e {
		if eim, ok := m[ei.Value]; ok {
			eim.requestTagCount++
			eim.requestTagIndex += ei.requestTagIndex
			continue
		}

		ei.requestTagCount = 1
		m[ei.Value] = ei
		u = append(u, ei)
	}

	return u
}

func mapEntries(e []*Entry) []string {
	v := make([]string, 0, len(e))
	for _, ei := range e {
		v = append(v, ei.Value)
	}

	return v
}

func (t *TagStash) getAll(tags []string) ([]*Entry, error) {
	entries, err := t.cache.Get(tags)
	if err != nil {
		return nil, err
	}

	notCached := setRequestIndex(tags, entries)

	stored, err := t.storage.Get(notCached)
	if err != nil {
		return nil, err
	}

	for _, e := range stored {
		if err := t.cache.Set(e); err != nil {
			return nil, err
		}
	}

	setRequestIndex(tags, stored)
	entries = append(entries, stored...)

	entries = uniqueValues(entries)
	sort.Sort(entrySort{tags, entries})
	return entries, nil
}

// Get returns the best matching value for a set of tags. When there are overlapping tags and values, it
// prioritizes first those values that match more tags from the arguments. When there are matches with the same
// number of matching tags, it prioritizes those that whose tag order matches the closer the order of the tags
// in the arguments. The tag order means the order of tags at the time of the definition (Set()).
func (t *TagStash) Get(tags ...string) (string, error) {
	entries, err := t.getAll(tags)
	if err != nil {
		return "", err
	}

	if len(entries) == 0 {
		return "", nil
	}

	v := mapEntries(entries[:1])
	return v[0], nil
}

// GetAll returns all matches for a set of tags, sorted by the same rules that are used for prioritization when
// calling Get().
func (t *TagStash) GetAll(tags ...string) ([]string, error) {
	entries, err := t.getAll(tags)
	if err != nil {
		return nil, err
	}

	return mapEntries(entries), nil
}

// Set stores tags associated with a value. The order of the tags is taken into account when there are
// overlapping matches during retrieval.
func (t *TagStash) Set(value string, tags ...string) error {
	for i, ti := range tags {
		e := &Entry{
			Value:    value,
			Tag:      ti,
			TagIndex: i,
		}

		if err := t.storage.Set(e); err != nil {
			return err
		}

		if err := t.cache.Set(e); err != nil {
			return err
		}
	}

	return nil
}

// Remove deletes a value-tag association.
func (t *TagStash) Remove(value string, tag string) error {
	e := &Entry{Value: value, Tag: tag}

	if err := t.cache.Remove(e); err != nil {
		return err
	}

	if err := t.storage.Remove(e); err != nil {
		return err
	}

	return nil
}

// Delete deletes all associations of a tag.
func (t *TagStash) Delete(tag string) error {
	if err := t.cache.Delete(tag); err != nil {
		return err
	}

	if err := t.storage.Delete(tag); err != nil {
		return err
	}

	return nil
}
