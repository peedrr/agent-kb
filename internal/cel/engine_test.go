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
