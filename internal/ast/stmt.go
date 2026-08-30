package ast

import "github.com/harshagw/viri/internal/token"

// Stmt is the statement interface.
type Stmt interface {
	Node
	stmtNode()
}

type ExprStmt struct {
	Expr Expr
}

func (*ExprStmt) stmtNode()                       {}
func (s *ExprStmt) GetPrimaryToken() *token.Token { return s.Expr.GetPrimaryToken() }

type PrintStmt struct {
	Expr Expr
}

func (*PrintStmt) stmtNode()                       {}
func (s *PrintStmt) GetPrimaryToken() *token.Token { return s.Expr.GetPrimaryToken() }

type VarDeclStmt struct {
	Name        *token.Token
	Type        TypeExpr // never nil: the grammar requires an annotation
	Initializer Expr
	Exported    bool
	IsConst     bool
}

func (*VarDeclStmt) stmtNode()                       {}
func (s *VarDeclStmt) GetPrimaryToken() *token.Token { return s.Name }

type BlockStmt struct {
	Statements []Stmt
	// LeftBrace is the opening '{', used as the position when the block is
	// empty and so has no statement to point at.
	LeftBrace *token.Token
}

func (*BlockStmt) stmtNode() {}
func (s *BlockStmt) GetPrimaryToken() *token.Token {
	if len(s.Statements) > 0 {
		if tok := s.Statements[0].GetPrimaryToken(); tok != nil {
			return tok
		}
	}
	return s.LeftBrace
}

type IfStmt struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (*IfStmt) stmtNode()                       {}
func (s *IfStmt) GetPrimaryToken() *token.Token { return s.Condition.GetPrimaryToken() }

type WhileStmt struct {
	Condition Expr
	Body      Stmt
}

func (*WhileStmt) stmtNode()                       {}
func (s *WhileStmt) GetPrimaryToken() *token.Token { return s.Condition.GetPrimaryToken() }

type ForStmt struct {
	Initializer Stmt
	Condition   Expr
	Increment   Expr
	Body        Stmt
	// Keyword is the 'for' itself, used as the position when every clause
	// is omitted (`for (;;) { ... }`).
	Keyword *token.Token
}

func (*ForStmt) stmtNode() {}
func (s *ForStmt) GetPrimaryToken() *token.Token {
	if s.Initializer != nil {
		if tok := s.Initializer.GetPrimaryToken(); tok != nil {
			return tok
		}
	}
	if s.Condition != nil {
		if tok := s.Condition.GetPrimaryToken(); tok != nil {
			return tok
		}
	}
	return s.Keyword
}

type BreakStmt struct {
	Keyword *token.Token
}

func (*BreakStmt) stmtNode()                       {}
func (s *BreakStmt) GetPrimaryToken() *token.Token { return s.Keyword }

type ContinueStmt struct {
	Keyword *token.Token
}

func (*ContinueStmt) stmtNode()                       {}
func (s *ContinueStmt) GetPrimaryToken() *token.Token { return s.Keyword }

type FunctionStmt struct {
	Name       *token.Token
	Params     []Param
	ReturnType TypeExpr // nil means the function returns nothing
	Body       *BlockStmt
	Exported   bool
}

func (*FunctionStmt) stmtNode()                       {}
func (s *FunctionStmt) GetPrimaryToken() *token.Token { return s.Name }

type ReturnStmt struct {
	Keyword *token.Token
	Value   Expr
}

func (*ReturnStmt) stmtNode()                       {}
func (s *ReturnStmt) GetPrimaryToken() *token.Token { return s.Keyword }

type ClassStmt struct {
	Name       *token.Token
	SuperClass *VariableExpr
	Fields     []*FieldDecl
	Methods    []*FunctionStmt
	Exported   bool
}

func (*ClassStmt) stmtNode()                       {}
func (s *ClassStmt) GetPrimaryToken() *token.Token { return s.Name }

type ImportStmt struct {
	Path  *token.Token
	Alias *token.Token
}

func (*ImportStmt) stmtNode()                       {}
func (s *ImportStmt) GetPrimaryToken() *token.Token { return s.Path }
