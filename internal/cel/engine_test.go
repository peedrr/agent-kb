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

// tripleHeadingsRule scans every triple of headings and its predicate is never
// satisfied, so its evaluation cost grows cubically with the heading count.
const tripleHeadingsRule = "page.ast.headings.exists(a, page.ast.headings.exists(b, page.ast.headings.exists(c, a.level == 0 && b.level == 0 && c.level == 0)))"

// tripleHeadingsVars builds a page whose heading list makes tripleHeadingsRule
// exhaustively expensive: n headings cost n*n*n evaluated triples.
func tripleHeadingsVars(n int) map[string]any {
	headings := make([]any, n)
	for i := range headings {
		headings[i] = map[string]any{"level": i%6 + 1}
	}
	return map[string]any{
		"page": map[string]any{
			"ast": map[string]any{"headings": headings},
		},
	}
}

func TestCompileRuleCostLimit(t *testing.T) {
	env, err := NewEnv()
	if err != nil {
		t.Fatalf("NewEnv: %v", err)
	}

	t.Run("pathological rule", func(t *testing.T) {
		prg, err := CompileRule(env, tripleHeadingsRule)
		if err != nil {
			t.Fatalf("CompileRule: %v", err)
		}

		evaluated := make(chan error, 1)
		go func() {
			_, err := Evaluate(context.Background(), prg, tripleHeadingsVars(2048), MaxCostLimit)
			evaluated <- err
		}()

		select {
		case err := <-evaluated:
			if err == nil {
				t.Fatal("rule evaluated without hitting the compiled cost budget")
			}
			if want := "exceeded compute budget"; err.Error() != want {
				t.Errorf("error = %q, want %q", err.Error(), want)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("rule did not stop at the cost budget: still evaluating after 10s")
		}
	})

	t.Run("tiny budget", func(t *testing.T) {
		ast, issues := env.Compile(tripleHeadingsRule)
		if issues != nil && issues.Err() != nil {
			t.Fatalf("Compile: %v", issues.Err())
		}

		const tinyCostLimit = 1
		prg, err := env.Program(ast, cel.CostLimit(tinyCostLimit))
		if err != nil {
			t.Fatalf("Program: %v", err)
		}

		_, err = Evaluate(context.Background(), prg, tripleHeadingsVars(16), tinyCostLimit)
		if err == nil {
			t.Fatal("expected error for cost limit exceeded")
		}
		want := "exceeded compute budget"
		if err.Error() != want {
			t.Errorf("error = %q, want %q", err.Error(), want)
		}
	})
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
