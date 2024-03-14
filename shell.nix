let
    nixpkgs = fetchTarball "https://github.com/NixOS/nixpkgs/tarball/nixos-unstable";
    pkgs = import nixpkgs { config = {}; overlays = []; };
in

pkgs.mkShell {
    packages = with pkgs; [
        go
    ];

    shellHook = ''
        go install golang.org/x/tools/gopls@latest
        go install golang.org/x/tools/cmd/goimports@latest
        export PATH="$PATH:$HOME/go/bin"
    '';
}

