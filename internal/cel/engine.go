package cel

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"reflect"
	"sync"
)

var programCache = sync.Map{}

// NewEnv creates a CEL environment pre-configured with native types for
// AST elements and variables for page context evaluation.
func NewEnv() (*cel.Env, error) {
	return cel.NewEnv(
		ext.NativeTypes(
			reflect.TypeOf(Heading{}),
			reflect.TypeOf(Link{}),
			reflect.TypeOf(CodeBlock{}),
		),
		cel.Variable("page", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("old_page", cel.NullableType(cel.MapType(cel.StringType, cel.DynType))),
		cel.Variable("now", cel.TimestampType),
	)
}

// CompileRule parses and compiles a CEL expression into an executable program.
// Compiled programs are cached in a thread-safe sync.Map to avoid re-compilation.
func CompileRule(env *cel.Env, expr string) (cel.Program, error) {
	if cached, ok := programCache.Load(expr); ok {
		return cached.(cel.Program), nil
	}

	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	prg, err := env.Program(ast, cel.CostTracking(nil))
	if err != nil {
		return nil, err
	}

	programCache.Store(expr, prg)
	return prg, nil
}
