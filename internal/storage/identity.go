// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package storage

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	yaml "github.com/goccy/go-yaml"

	"github.com/peedrr/agent-kb/internal/path"
)

// Environment variables that name the commit identity of an invocation. They
// outrank everything a base records and everything git configures, so an agent
// attributes its own commits without touching either.
const (
	AuthorNameEnv  = "AKB_AUTHOR_NAME"
	AuthorEmailEnv = "AKB_AUTHOR_EMAIL"
)

// DefaultIdentity is the commit identity akb records for a base when neither
// the invocation, nor the environment, nor git names one.
var DefaultIdentity = Identity{Name: "agent-kb", Email: "agent@agent-kb"}

// ErrNoCommitIdentity reports that no commit identity resolves for a base:
// AKB_AUTHOR_NAME/AKB_AUTHOR_EMAIL name none, its akb.yaml names none, and git's
// own configuration names none. Its message carries both remedies, and the CLI
// maps it to the usage exit code, since it is a configuration fault the caller
// can fix rather than a failed operation. It is the failure of the identity
// preflight: no file of the operation is written before it is reported.
var ErrNoCommitIdentity = errors.New("no commit identity: set AKB_AUTHOR_NAME and AKB_AUTHOR_EMAIL, or git-author and git-email in .agent-kb/akb.yaml")

// Identity is a commit identity: the name and the email address a commit
// records.
type Identity struct {
	Name  string
	Email string
}

// IdentitySource names where the identity of a new base came from.
type IdentitySource int

const (
	// IdentityFromFlag is the --author-name/--author-email flags of `akb init`.
	IdentityFromFlag IdentitySource = iota
	// IdentityFromEnv is the AKB_AUTHOR_NAME/AKB_AUTHOR_EMAIL environment.
	IdentityFromEnv
	// IdentityFromGitConfig is git's merged user.name/user.email configuration.
	IdentityFromGitConfig
	// IdentityFromDefault is nothing naming an identity, so DefaultIdentity
	// applies.
	IdentityFromDefault
)

// ResolveIdentity returns the identity akb supplies to the commits of the base
// rooted at kbRoot, resolved in precedence order: AKB_AUTHOR_NAME and
// AKB_AUTHOR_EMAIL, then git-author and git-email in the base's akb.yaml, then
// git's own resolution — GIT_AUTHOR_* in the environment, then repository and
// global configuration. A nil identity means akb supplies nothing: git resolves
// the identity itself, so an identity the caller exports through GIT_AUTHOR_*
// and GIT_COMMITTER_* stays the one its commits record.
//
// A field no akb source names is filled from git's own resolution, so a
// returned identity is complete. ErrNoCommitIdentity reports that no identity
// resolves at all: git would refuse the commit for want of one, so callers
// report it before writing anything.
func ResolveIdentity(kbRoot string) (*Identity, error) {
	configured, err := configuredIdentity(kbRoot)
	if err != nil {
		return nil, err
	}

	name := firstNonEmpty(strings.TrimSpace(os.Getenv(AuthorNameEnv)), configured.Name)
	email := firstNonEmpty(strings.TrimSpace(os.Getenv(AuthorEmailEnv)), configured.Email)

	if name == "" && email == "" {
		// Git resolves the identity on its own, under the caller's environment
		// and configuration. Both the author and the committer identity have to
		// resolve: a commit records both, and akb passes nothing for either.
		if _, authorOK := gitIdent(kbRoot, "GIT_AUTHOR_IDENT"); !authorOK {
			return nil, ErrNoCommitIdentity
		}
		if _, committerOK := gitIdent(kbRoot, "GIT_COMMITTER_IDENT"); !committerOK {
			return nil, ErrNoCommitIdentity
		}
		return nil, nil
	}

	// A field no akb source names comes from git. The identity akb returns is
	// complete, and it exports both the author and the committer, so a commit
	// akb supplies the identity for cannot fail for want of a committer one.
	if name == "" || email == "" {
		native := nativeIdentity(kbRoot)
		name, email = firstNonEmpty(name, native.Name), firstNonEmpty(email, native.Email)
	}
	if name == "" || email == "" {
		return nil, ErrNoCommitIdentity
	}
	return &Identity{Name: name, Email: email}, nil
}

// ResolveInitIdentity returns the commit identity of a base akb is creating at
// kbDir, resolved in precedence order: the --author-name and --author-email
// flags of `akb init`, then AKB_AUTHOR_NAME and AKB_AUTHOR_EMAIL, then git's
// merged user.name and user.email, then DefaultIdentity. A field the source it
// is read from leaves empty is filled from the next source down the chain.
//
// The source names the highest-precedence source that contributed, which
// decides what the base records: an identity that a flag, the environment, or
// git config named belongs to the invocation or to the machine and is never
// written to akb.yaml, while the default is recorded so that a base cloned to a
// machine without a git identity still commits.
func ResolveInitIdentity(kbDir, flagName, flagEmail string) (Identity, IdentitySource) {
	flagName, flagEmail = strings.TrimSpace(flagName), strings.TrimSpace(flagEmail)
	envName, envEmail := strings.TrimSpace(os.Getenv(AuthorNameEnv)), strings.TrimSpace(os.Getenv(AuthorEmailEnv))
	// Git names an identity only when it can attribute a commit with one: a
	// configuration it cannot form a complete identity from names none, and the
	// default applies instead, so the new base can always commit.
	gitIdentity, gitNamed := gitIdent(kbDir, "GIT_AUTHOR_IDENT")

	identity := Identity{
		Name:  firstNonEmpty(flagName, envName, gitIdentity.Name, DefaultIdentity.Name),
		Email: firstNonEmpty(flagEmail, envEmail, gitIdentity.Email, DefaultIdentity.Email),
	}

	source := IdentityFromDefault
	switch {
	case flagName != "" || flagEmail != "":
		source = IdentityFromFlag
	case envName != "" || envEmail != "":
		source = IdentityFromEnv
	case gitNamed:
		source = IdentityFromGitConfig
	}
	return identity, source
}

// identityConfig is the part of akb.yaml that names the base's commit identity.
// Only these two keys are read: the identity has to resolve for the first
// commit of a base, whose config may carry nothing else yet, and for a base
// whose config was never written at all.
type identityConfig struct {
	GitAuthor string `yaml:"git-author"`
	GitEmail  string `yaml:"git-email"`
}

// configuredIdentity returns the commit identity the base's akb.yaml names. A
// base without a config, or one that names no identity, configures none.
func configuredIdentity(kbRoot string) (Identity, error) {
	data, err := os.ReadFile(path.ConfigPath(kbRoot))
	if err != nil {
		if os.IsNotExist(err) {
			return Identity{}, nil
		}
		return Identity{}, fmt.Errorf("read base config: %w", err)
	}

	var cfg identityConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Identity{}, fmt.Errorf("parse base config: %w", err)
	}
	return Identity{Name: strings.TrimSpace(cfg.GitAuthor), Email: strings.TrimSpace(cfg.GitEmail)}, nil
}

// nativeIdentity returns the identity git resolves on its own for the
// repository of kbRoot. `git var GIT_AUTHOR_IDENT` is the identity git itself
// would record, with GIT_AUTHOR_* from the environment and the merged
// repository and global configuration applied; a field it cannot name is left
// empty. The individual configuration values are the fallback, because git
// refuses to name an identity whose name is empty while such a repository still
// commits under a name akb supplies.
func nativeIdentity(kbRoot string) Identity {
	if ident, ok := gitIdent(kbRoot, "GIT_AUTHOR_IDENT"); ok {
		return ident
	}
	return Identity{
		Name:  gitConfigValue(kbRoot, "user.name"),
		Email: gitConfigValue(kbRoot, "user.email"),
	}
}

// gitIdent returns the identity `git var` reports for name — GIT_AUTHOR_IDENT or
// GIT_COMMITTER_IDENT — which is the identity a commit would record with akb's
// environment and the merged configuration applied. It is absent when git
// cannot name one.
func gitIdent(kbRoot, name string) (Identity, bool) {
	cmd := exec.Command("git", "var", name) //nolint:gosec // launching trusted git binary with controlled args
	cmd.Dir = kbRoot
	out, err := cmd.Output()
	if err != nil {
		return Identity{}, false
	}
	return parseIdent(string(out))
}

// parseIdent parses the identity line git prints, "<name> <<email>> <timestamp>
// <timezone>".
func parseIdent(out string) (Identity, bool) {
	line := strings.TrimSpace(out)
	emailEnd := strings.LastIndex(line, ">")
	if emailEnd < 0 {
		return Identity{}, false
	}
	emailStart := strings.LastIndex(line[:emailEnd], "<")
	if emailStart < 0 {
		return Identity{}, false
	}

	identity := Identity{
		Name:  strings.TrimSpace(line[:emailStart]),
		Email: strings.TrimSpace(line[emailStart+1 : emailEnd]),
	}
	if identity.Name == "" || identity.Email == "" {
		return Identity{}, false
	}
	return identity, true
}

// gitConfigValue returns the merged configuration value of key, or "" when git
// sets none. An empty value counts as unset: it names no identity.
func gitConfigValue(kbRoot, key string) string {
	cmd := exec.Command("git", "config", key) //nolint:gosec // launching trusted git binary with controlled args
	cmd.Dir = kbRoot
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// commitEnv returns the environment of a git process that commits under
// identity: both the author and the committer, so the commit is attributed to
// akb's identity whatever the repository configures. A nil identity inherits
// akb's environment untouched, which lets git resolve the identity itself.
func commitEnv(identity *Identity) []string {
	if identity == nil {
		return nil
	}
	return envWithOverrides([]string{
		"GIT_AUTHOR_NAME=" + identity.Name,
		"GIT_AUTHOR_EMAIL=" + identity.Email,
		"GIT_COMMITTER_NAME=" + identity.Name,
		"GIT_COMMITTER_EMAIL=" + identity.Email,
	})
}

// envWithOverrides returns akb's environment with the given NAME=value
// overrides applied. exec passes the slice to the child verbatim and a variable
// that appears twice is read from its first entry, so the overridden names are
// dropped from the inherited environment instead of being shadowed by a second
// copy of them.
func envWithOverrides(overrides []string) []string {
	overridden := make(map[string]bool, len(overrides))
	for _, override := range overrides {
		key, _, _ := strings.Cut(override, "=")
		overridden[key] = true
	}

	env := make([]string, 0, len(os.Environ())+len(overrides))
	for _, variable := range os.Environ() {
		key, _, _ := strings.Cut(variable, "=")
		if overridden[key] {
			continue
		}
		env = append(env, variable)
	}
	return append(env, overrides...)
}

// firstNonEmpty returns the first value that is not empty.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
