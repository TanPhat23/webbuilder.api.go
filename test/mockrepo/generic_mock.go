package test

// GenericMock is a reusable base mock type using a function map.
type GenericMock struct {
	funcs map[string]any
}

// Set registers a function with a given name and returns the mock for chaining.
func (g *GenericMock) Set(name string, fn any) *GenericMock {
	if g.funcs == nil {
		g.funcs = make(map[string]any)
	}
	g.funcs[name] = fn
	return g
}

// Get retrieves a registered function by name.
func (g *GenericMock) Get(name string) any {
	if g.funcs == nil {
		return nil
	}
	return g.funcs[name]
}