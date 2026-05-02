package cel

import "testing"

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
