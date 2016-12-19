package tagstash

type storage struct{}

func newStorage(StorageOptions) *storage { return &storage{} }

func (s *storage) Get([]string) ([]*Entry, error) { return nil, nil }
func (s *storage) Set(*Entry) error               { return nil }
func (s *storage) Remove(*Entry) error            { return nil }
func (s *storage) Delete(string) error            { return nil }
func (s *storage) Close()                         {}
