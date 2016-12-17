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

// Logger objects are used to log runtime diagnostic messages. They are used report errors on read operations,
// when the read failed for individual items, but it was able to continue with others.
type Logger interface {

	// Error outputs messages of error severity.
	Error(a ...interface{})
}

type entrySort struct {
	tags    []string
	entries []*Entry
}

// TagStash is used to store tags associated with values and return the best matching value for a set of query
// tags.
type TagStash struct {
	cache, storage Storage
	logger         Logger
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

func (t *TagStash) getAll(tags []string) []*Entry {
	entries, err := t.cache.Get(tags)
	if err != nil {
		t.logger.Error("error while accessing cache", err)
	}

	notCached := setRequestIndex(tags, entries)

	stored, err := t.storage.Get(notCached)
	if err != nil {
		t.logger.Error("error while accessing storage", err)
	}

	for _, e := range stored {
		if err := t.cache.Set(e); err != nil {
			t.logger.Error("error while caching entry", e.Value, e.Tag, err)
		}
	}

	setRequestIndex(tags, stored)
	entries = append(entries, stored...)

	entries = uniqueValues(entries)
	sort.Sort(entrySort{tags, entries})
	return entries
}

// Get returns the best matching value for a set of tags. When there are overlapping tags and values, it
// prioritizes first those values that match more tags from the arguments. When there are matches with the same
// number of matching tags, it prioritizes those that whose tag order matches the closer the order of the tags
// in the arguments. The tag order means the order of tags at the time of the definition (Set()).
func (t *TagStash) Get(tags []string) (string, bool) {
	entries := t.getAll(tags)
	if len(entries) == 0 {
		return "", false
	}

	v := mapEntries(entries[:1])
	return v[0], true
}

// GetAll returns all matches for a set of tags, sorted by the same rules that are used for prioritization when
// calling Get().
func (t *TagStash) GetAll(tags []string) []string {
	entries := t.getAll(tags)
	return mapEntries(entries)
}

// Set stores tags associated with a value. The order of the tags is taken into account when there are
// overlapping matches during retrieval.
func (t *TagStash) Set(value string, tags []string) error {
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
