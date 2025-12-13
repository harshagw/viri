package ast

type Module struct {
	Path       string
	Imports    []*ImportStmt
	Statements []Stmt
}

func NewModule(path string, imports []*ImportStmt, statements []Stmt) *Module {
	return &Module{
		Path:       path,
		Imports:    imports,
		Statements: statements,
	}
}

func (m *Module) GetAllStatements() []Stmt {
	all := make([]Stmt, 0, len(m.Imports)+len(m.Statements))
	for _, importStmt := range m.Imports {
		all = append(all, importStmt)
	}
	all = append(all, m.Statements...)
	return all
}
