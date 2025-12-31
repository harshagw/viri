package ast

import "github.com/harshagw/viri/internal/token"

// Node is the base interface for all AST nodes.
type Node interface {
	GetPrimaryToken() *token.Token
}

func GetNodeFilePath(node Node) string {
	if tok := node.GetPrimaryToken(); tok != nil && tok.FilePath != nil {
		return *tok.FilePath
	}
	return ""
}
