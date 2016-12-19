package tagstash

import "errors"

type mockStorage struct {
	entries  []*Entry
	failNext bool
}

var errForgedError = errors.New("forged")

func (s *mockStorage) fail() error {
	if s.failNext {
		s.failNext = false
		return errForgedError
	}

	return nil
}

func (s *mockStorage) Get(tags []string) ([]*Entry, error) {
	if err := s.fail(); err != nil {
		return nil, err
	}

	var entries []*Entry
	for _, e := range s.entries {
		for _, t := range tags {
			if t == e.Tag {
				entries = append(entries, e)
			}
		}
	}

	return entries, nil
}

func (s *mockStorage) Set(e *Entry) error {
	if err := s.fail(); err != nil {
		return err
	}

	for _, ei := range s.entries {
		if ei.Tag == e.Tag && ei.Value == e.Value {
			ei.TagIndex = e.TagIndex
			return nil
		}
	}

	s.entries = append(s.entries, e)
	return nil
}

func (s *mockStorage) Remove(e *Entry) error {
	if err := s.fail(); err != nil {
		return err
	}

	for i, ei := range s.entries {
		if ei.Tag == e.Tag && ei.Value == e.Value {
			s.entries[len(s.entries)-1], s.entries =
				nil, append(s.entries[:i], s.entries[i+1:]...)
			return nil
		}
	}

	return nil
}

func (s *mockStorage) Delete(tag string) error {
	if err := s.fail(); err != nil {
		return err
	}

	next := make([]*Entry, 0, len(s.entries))
	for _, e := range s.entries {
		if e.Tag != tag {
			next = append(next, e)
		}
	}

	s.entries = next
	return nil
}

func (s *mockStorage) Close() {}
