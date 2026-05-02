package cel

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"reflect"
)

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
