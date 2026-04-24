package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

var akbBin string

func TestMain(m *testing.M) {
	// Get current working directory - tests run from test/ directory
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// test/ is under project root, so go up one level
	projRoot := filepath.Dir(cwd)

	tmpDir, err := os.MkdirTemp("", "akb-inttest")
	if err != nil {
		panic(err)
	}
	akbBin = filepath.Join(tmpDir, "akb")

	buildCmd := exec.Command( //nolint:gosec // test helper launching akb binary
		"go", "build", "-o", akbBin, "./cmd/akb/")
	buildCmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	buildCmd.Dir = projRoot
	if err := buildCmd.Run(); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func Test(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(env *testscript.Env) error {
			binDir := filepath.Dir(akbBin)
			for i, v := range env.Vars {
				if strings.HasPrefix(v, "PATH=") {
					env.Vars[i] = "PATH=" + binDir + string(os.PathListSeparator) + v[5:]
					break
				}
			}
			// Create a unique HOME directory per test to isolate the KB
			// registry file and avoid parallel test conflicts
			testHome := filepath.Join(env.WorkDir, "home")
			if err := os.MkdirAll(testHome, 0750); err != nil {
				return err
			}
			env.Vars = append(env.Vars, "HOME="+testHome)
			return nil
		},
	})
}
