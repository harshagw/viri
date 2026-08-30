package objects

// FunctionType distinguishes the syntactic forms a function can take. It is
// used to build parser diagnostics ("Expect '(' after named function.").
type FunctionType int

const (
	FunctionTypeAnonymous FunctionType = iota
	FunctionTypeNamed
)

func (ft FunctionType) String() string {
	switch ft {
	case FunctionTypeAnonymous:
		return "anonymous"
	case FunctionTypeNamed:
		return "named"
	}
	return "unknown function type"
}
