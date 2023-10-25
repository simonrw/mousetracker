{
  description = "Flake utils demo";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-23.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [
        ];

        pkgs = import nixpkgs {
          inherit overlays system;
        };
      in
      {
        packages.default = import ./default.nix { inherit pkgs; };
        devShells.default = pkgs.mkShell {
          inputsFrom = [
            self.outputs.packages.${system}.default
          ];
        };
      }
    );
}
