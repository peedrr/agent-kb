// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/peedrr/agent-kb/internal/path"
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
			// HOME is the scenario work dir itself: the scenarios cannot read
			// the developer's configuration, and a neighborhood walk from any
			// directory inside the scenario is bounded there instead of
			// climbing past it to the filesystem root.
			env.Vars = append(env.Vars, "HOME="+env.WorkDir)
			// Ambient git identity, injected through GIT_CONFIG_* entries that
			// outrank system and global configuration: scenarios that shell out
			// to raw `git commit` (template seeding, merge setup) must not
			// depend on the machine's own identity. NixOS hosts resolve one
			// from system-level config even with HOME moved; CI runners have
			// none, which is how machine-dependent scenarios pass locally and
			// fail in CI. Scenarios that exercise identity resolution itself
			// override these same slots with `env` (which runs after Setup) or
			// empty GIT_CONFIG_VALUE_0/1 to simulate a host without identity.
			env.Vars = append(env.Vars,
				"GIT_CONFIG_COUNT=2",
				"GIT_CONFIG_KEY_0=user.name",
				"GIT_CONFIG_VALUE_0=akb-test",
				"GIT_CONFIG_KEY_1=user.email",
				"GIT_CONFIG_VALUE_1=akb-test@localhost",
			)
			// Every scenario runs its commands from the root of the KB it
			// created, so a relative selection resolves against that working
			// directory. Scenarios that need no KB clear the variable.
			env.Vars = append(env.Vars, path.KBEnvVar+"=.")
			return nil
		},
	})
}
