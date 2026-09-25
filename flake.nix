{
  description = "agent-kb: the knowledge base your AI agents want";

  inputs.nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/0.1"; # unstable Nixpkgs

  outputs =
    { self, ... }@inputs:

    let
      # Single source of truth: the VERSION file at the repo root. The justfile
      # derives from the same file, so the two build paths cannot drift; the
      # human-facing copies (AGENTS.md, CHANGELOG.md, the git tag) are checked
      # by scripts/check-version-lockstep.sh. NOTE: readFile only sees
      # git-tracked files in a flake source, so VERSION must be committed
      # before `nix build` will pick up a bump.
      akbVersion = inputs.nixpkgs.lib.removeSuffix "\n" (builtins.readFile ./VERSION);

      goVersion = 26; # Change this to update the whole stack

      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forEachSupportedSystem =
        f:
        inputs.nixpkgs.lib.genAttrs supportedSystems (
          system:
          f {
            inherit system;
            pkgs = import inputs.nixpkgs {
              inherit system;
              overlays = [ inputs.self.overlays.default ];
            };
          }
        );
    in
    {
      overlays.default = final: prev: {
        go = final."go_1_${toString goVersion}";
      };

      devShells = forEachSupportedSystem (
        { pkgs, system }:
        {
          default = pkgs.mkShellNoCC {
            # The SQLite driver is pure Go; the integration test harness
            # also forces CGO_ENABLED=0, so the dev shell matches it.
            env.CGO_ENABLED = "0";

            packages = with pkgs; [
              # go (version is specified by overlay)
              go

              # go lsp
              gopls

              # go debugger
              delve

              # goimports, godoc, etc.
              gotools

              # https://github.com/golangci/golangci-lint
              golangci-lint

              # Go vulnerability database
              govulncheck

              # task runner (build/test/lint/release recipes)
              just

              self.formatter.${system}
              self.nixLsp.${system}

              tmux
            ];
          };
        }
      );

      packages = forEachSupportedSystem (
        { pkgs, ... }:
        {
          default = pkgs.buildGoModule {
            pname = "agent-kb";
            version = akbVersion;
            src = ./.;

            # To update the hash:
            # 1. Set vendorHash = lib.fakeHash;
            # 2. Run 'nix build'
            # 3. Copy the 'got:' hash from the error message
            vendorHash = "sha256-nPgj+yPf2JEhV4l+TlCrwrdvi/8QV0YSiUWZYHJFwMk=";

            subPackages = [ "cmd/akb" ];

            ldflags = [
              "-X main.version=${akbVersion}"
            ];

            # Ensure we use the Go version specified in the flake
            nativeBuildInputs = [ pkgs.go ];

            # Required for integration tests that run git commands
            nativeCheckInputs = [ pkgs.git ];

            meta = {
              description = "Agent KB: a git-tracked knowledge base CLI with SQLite search, link graph, and CEL validation";
              homepage = "https://github.com/peedrr/agent-kb";
              license = pkgs.lib.licenses.mpl20;
              mainProgram = "akb";
            };
          };
        }
      );

      apps = forEachSupportedSystem (
        { system, ... }:
        {
          default = {
            type = "app";
            program = "${self.packages.${system}.default}/bin/akb";
          };
        }
      );

      formatter = forEachSupportedSystem ({ pkgs, ... }: pkgs.nixfmt);
      nixLsp = forEachSupportedSystem ({ pkgs, ... }: pkgs.nixd);
    };
}
