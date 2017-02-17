package irgen

import gooryvalues "github.com/bongo227/goory/value"

type Scope struct {
	parentScope  *Scope
	currentScope map[string]Item
}

type Item struct {
	Key   string
	Value gooryvalues.Value
}

func NewScope() *Scope {
	return &Scope{
		parentScope:  nil,
		currentScope: make(map[string]Item),
	}
}

func (s *Scope) Push() *Scope {
	return &Scope{
		parentScope:  s,
		currentScope: make(map[string]Item),
	}
}

// Add adds a value to the current scope
func (s *Scope) Add(key string, value gooryvalues.Value) {
	s.currentScope[key] = Item{key, value}
}

// GetLocal returns the item and true if the key is in local scope, otherwise false
func (s *Scope) GetLocal(key string) (Item, bool) {
	if item, ok := s.currentScope[key]; ok {
		return item, true
	}

	return Item{}, false
}

// Get returns the item and true if the key is in local (or parent scope), otherwise
// false
func (s *Scope) Get(key string) (Item, bool) {
	for s.parentScope != nil {
		if item, ok := s.currentScope[key]; ok {
			return item, true
		}

		s = s.parentScope
	}

	// Check root scope (parentScope will be nil)
	return s.GetLocal(key)
}
