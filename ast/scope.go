package ast

type Scope struct {
	parent *Scope
	scope  map[string]Node
}

// NewScope creates a new blank scope
func NewScope() *Scope {
	return &Scope{
		parent: nil,
		scope:  make(map[string]Node),
	}
}

// Enter creates a new inner scope
func (s *Scope) Enter() *Scope {
	child := NewScope()
	child.parent = s
	return child
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
		currentScope = currentScope.parent
	}

	return nil
}

func (s *Scope) Replace(name string, node Node) bool {
	currentScope := s
	for currentScope != nil {
		if currentScope.scope[name] != nil {
			currentScope.scope[name] = node
			return true
		}
		currentScope = currentScope.parent
	}

	return false
}

// Exit returns the outer scope
func (s *Scope) Exit() *Scope {
	return s.parent
}
