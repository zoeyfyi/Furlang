package ast

type Scope struct {
	outer *Scope
	scope map[string]Node
}

// NewScope creates a new blank scope
func NewScope() *Scope {
	return &Scope{
		outer: nil,
		scope: make(map[string]Node),
	}
}

// Enter creates a new inner scope
func (s *Scope) Enter() *Scope {
	inner := NewScope()
	inner.outer = s
	return inner
}

// Insert adds a new declaration to the current scope
func (s *Scope) Insert(name string, node Node) {
	s.scope[name] = node
}

// Lookup returns the node in the most inner scope
func (s *Scope) Lookup(name string) Node {
	currentScope := s
	for currentScope != nil {
		if currentScope.scope[name] != nil {
			return currentScope.scope[name]
		}
		currentScope = currentScope.outer
	}

	return nil
}

// Exit returns the outer scope
func (s *Scope) Exit() *Scope {
	return s.outer
}
