package cel

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter"
)

func TestNewEnv(t *testing.T) {
	env, err := NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	ast, issues := env.Compile("page.ast.headings.exists(h, h.level == 2)")
	if issues != nil && issues.Err() != nil {
		t.Fatalf("compile heading rule: %v", issues.Err())
	}
	if ast == nil {
		t.Fatal("ast is nil")
	}
}

func TestCompileRule(t *testing.T) {
	env, err := NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	expr := "page.frontmatter.type == 'adr'"
	prg1, err := CompileRule(env, expr)
	if err != nil {
		t.Fatalf("first compile: %v", err)
	}
	if prg1 == nil {
		t.Fatal("program is nil")
	}

	prg2, err := CompileRule(env, expr)
	if err != nil {
		t.Fatalf("second compile: %v", err)
	}
	if prg2 != prg1 {
		t.Error("second compile did not return cached program")
	}
}

func TestCompileRuleError(t *testing.T) {
	env, err := NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	_, err = CompileRule(env, "invalid syntax here !!!")
	if err == nil {
		t.Fatal("expected compile error for invalid syntax")
	}
}

func TestEvaluate(t *testing.T) {
	env, err := NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	prg, err := CompileRule(env, "page.frontmatter.type == 'adr'")
	if err != nil {
		t.Fatalf("CompileRule: %v", err)
	}

	vars := map[string]any{
		"page": map[string]any{
			"frontmatter": map[string]any{"type": "adr"},
		},
	}

	result, err := Evaluate(context.Background(), prg, vars, 1000)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if result != types.True {
		t.Errorf("result = %v, want true", result)
	}
}

func TestEvaluateFalse(t *testing.T) {
	env, err := NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	prg, err := CompileRule(env, "page.frontmatter.type == 'adr'")
	if err != nil {
		t.Fatalf("CompileRule: %v", err)
	}

	vars := map[string]any{
		"page": map[string]any{
			"frontmatter": map[string]any{"type": "note"},
		},
	}

	result, err := Evaluate(context.Background(), prg, vars, 1000)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if result != types.False {
		t.Errorf("result = %v, want false", result)
	}
}

func TestCostLimit(t *testing.T) {
	env, err := NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	ast, issues := env.Compile("1 + 1")
	if issues != nil && issues.Err() != nil {
		t.Fatalf("Compile: %v", issues.Err())
	}

	prg, err := env.Program(ast, cel.CostLimit(0))
	if err != nil {
		t.Fatalf("Program: %v", err)
	}

	_, err = Evaluate(context.Background(), prg, map[string]any{}, 0)
	if err == nil {
		t.Fatal("expected error for cost limit exceeded")
	}
	want := "exceeded compute budget"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

type panicProgram struct {
	panicWith any
}

func (p panicProgram) Eval(any) (ref.Val, *cel.EvalDetails, error) {
	panic(p.panicWith)
}

func (p panicProgram) ContextEval(_ context.Context, _ any) (ref.Val, *cel.EvalDetails, error) {
	panic(p.panicWith)
}

func TestHasOnMaps(t *testing.T) {
	env, err := NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	prg, err := CompileRule(env, "has(page.frontmatter.type)")
	if err != nil {
		t.Fatalf("CompileRule: %v", err)
	}

	vars := map[string]any{
		"page": map[string]any{
			"frontmatter": map[string]any{"type": "adr"},
		},
	}

	result, err := Evaluate(context.Background(), prg, vars, 1000)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if result != types.True {
		t.Errorf("has(existing) = %v, want true", result)
	}

	prg2, err := CompileRule(env, "has(page.frontmatter.nonexistent)")
	if err != nil {
		t.Fatalf("CompileRule: %v", err)
	}

	result2, err := Evaluate(context.Background(), prg2, vars, 1000)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if result2 != types.False {
		t.Errorf("has(missing) = %v, want false", result2)
	}
}

func TestTypeError(t *testing.T) {
	env, err := NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	_, err = CompileRule(env, "page.frontmatter.type.size() == true")
	if err == nil {
		t.Fatal("expected compile error for type mismatch")
	}
	if !strings.Contains(err.Error(), "overload") {
		t.Errorf("error = %q, want actionable type error containing 'overload'", err.Error())
	}
}

func TestNowVariable(t *testing.T) {
	env, err := NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	prg, err := CompileRule(env, "now > timestamp('2000-01-01T00:00:00Z')")
	if err != nil {
		t.Fatalf("CompileRule: %v", err)
	}

	result, err := Evaluate(context.Background(), prg, map[string]any{
		"now": time.Now(),
	}, 1000)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if result != types.True {
		t.Errorf("now > 2000 = %v, want true", result)
	}
}

func TestPanicRecovery(t *testing.T) {
	t.Run("cost limit panic", func(t *testing.T) {
		prg := panicProgram{panicWith: interpreter.EvalCancelledError{
			Cause:   interpreter.CostLimitExceeded,
			Message: "operation cancelled: actual cost limit exceeded",
		}}
		_, err := Evaluate(context.Background(), prg, map[string]any{}, 100)
		if err == nil {
			t.Fatal("expected error")
		}
		want := "exceeded compute budget"
		if err.Error() != want {
			t.Errorf("error = %q, want %q", err.Error(), want)
		}
	})

	t.Run("unknown panic", func(t *testing.T) {
		prg := panicProgram{panicWith: "something went wrong"}
		_, err := Evaluate(context.Background(), prg, map[string]any{}, 100)
		if err == nil {
			t.Fatal("expected error")
		}
		want := "internal CEL error"
		if err.Error() != want {
			t.Errorf("error = %q, want %q", err.Error(), want)
		}
	})
}
