package runtime

// Env is a single scope frame in the scope chain.
type Env struct {
	vars   map[string]*Value
	parent *Env
}

func NewEnv(parent *Env) *Env {
	return &Env{vars: make(map[string]*Value), parent: parent}
}

// Define creates a new variable in the current scope.
func (e *Env) Define(name string, val Value) {
	e.vars[name] = &val
}

// Get walks up the scope chain to find a variable.
func (e *Env) Get(name string) Value {
	if v, ok := e.vars[name]; ok {
		return *v
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	panic("undefined variable: " + name)
}

// Set walks up the scope chain to update an existing variable.
func (e *Env) Set(name string, val Value) {
	if _, ok := e.vars[name]; ok {
		e.vars[name] = &val
		return
	}
	if e.parent != nil {
		e.parent.Set(name, val)
		return
	}
	panic("assignment to undefined variable: " + name)
}

// SetGlobal writes directly to the root scope.
func (e *Env) SetGlobal(name string, val Value) {
	root := e.root()
	if _, ok := root.vars[name]; ok {
		root.vars[name] = &val
		return
	}
	panic("assignment to undefined global: " + name)
}

// GetGlobal reads from the root scope.
func (e *Env) GetGlobal(name string) Value {
	root := e.root()
	if v, ok := root.vars[name]; ok {
		return *v
	}
	panic("undefined global variable: " + name)
}

func (e *Env) root() *Env {
	cur := e
	for cur.parent != nil {
		cur = cur.parent
	}
	return cur
}
