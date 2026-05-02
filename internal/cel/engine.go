package cel

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/ext"
	"github.com/google/cel-go/interpreter"
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

// Evaluate runs a compiled CEL program with the given variables and context.
// It recovers from panics during evaluation, translating cost-limit-exceeded
// errors into a structured "exceeded compute budget" error.
func Evaluate(ctx context.Context, prg cel.Program, vars map[string]any, costLimit uint64) (ref.Val, error) {
	var result ref.Val
	var evalErr error

	func() {
			defer func() {
				if r := recover(); r != nil {
					rerr, ok := r.(error)
					if !ok {
						evalErr = fmt.Errorf("internal CEL error")
						return
					}
					var cancelled interpreter.EvalCancelledError
					if errors.As(rerr, &cancelled) {
						evalErr = fmt.Errorf("exceeded compute budget")
						return
					}
					evalErr = fmt.Errorf("internal CEL error")
				}
			}()

		result, _, evalErr = prg.ContextEval(ctx, vars)
	}()

	if evalErr != nil {
		var cancelled interpreter.EvalCancelledError
		if errors.As(evalErr, &cancelled) {
			evalErr = fmt.Errorf("exceeded compute budget")
		}
		return nil, evalErr
	}

	return result, nil
}
