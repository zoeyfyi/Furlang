package irgen

import gooryvalues "github.com/bongo227/goory/value"
import "github.com/bongo227/goory"

type Scope struct {
	parentScope *Scope
	scope       *scopes
}

type scopes struct {
	functions map[string]*goory.Function
	varibles  map[string]gooryvalues.Value
}

func NewScope() *Scope {
	return &Scope{
		parentScope: nil,
		scope: &scopes{
			functions: make(map[string]*goory.Function),
			varibles:  make(map[string]gooryvalues.Value),
		},
	}
}

func (s *Scope) Push() *Scope {
	return &Scope{
		parentScope: s,
		scope: &scopes{
			functions: make(map[string]*goory.Function),
			varibles:  make(map[string]gooryvalues.Value),
		},
	}
}

// AddVar adds a value to the current scope
func (s *Scope) AddVar(key string, value gooryvalues.Value) {
	s.scope.varibles[key] = value
}

// GetLocalVar returns the item and true if the key is in local scope, otherwise false
func (s *Scope) GetLocalVar(key string) (gooryvalues.Value, bool) {
	if item, ok := s.scope.varibles[key]; ok {
		return item, true
	}

	return nil, false
}

// GetVar returns the item and true if the key is in local (or parent scope),
// otherwise false
func (s *Scope) GetVar(key string) (gooryvalues.Value, bool) {
	// TODO: do this with a map
	if key == "true" {
		return goory.Constant(goory.BoolType(), true), true
	} else if key == "false" {
		return goory.Constant(goory.BoolType(), false), true
	}

	for s.parentScope != nil {
		if item, ok := s.scope.varibles[key]; ok {
			return item, true
		}

		s = s.parentScope
	}

	// Check root scope (parentScope will be nil)
	return s.GetLocalVar(key)
}

// AddFunction adds a value to the current scope
func (s *Scope) AddFunction(key string, function *goory.Function) {
	s.scope.functions[key] = function
}

// GetLocalFunction returns the item and true if the key is in local scope, otherwise false
func (s *Scope) GetLocalFunction(key string) (*goory.Function, bool) {
	if item, ok := s.scope.functions[key]; ok {
		return item, true
	}

	return nil, false
}

// GetFunction returns the item and true if the key is in local (or parent scope),
// otherwise false
func (s *Scope) GetFunction(key string) (*goory.Function, bool) {
	for s.parentScope != nil {
		if item, ok := s.scope.functions[key]; ok {
			return item, true
		}

		s = s.parentScope
	}

	// Check root scope (parentScope will be nil)
	return s.GetLocalFunction(key)
}
