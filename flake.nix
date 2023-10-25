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
        nixosModules.default = { config, lib, ... }: 
        with lib;
        let
          cfg = config.mousetracker;
        in
        {
          options = {
            enable = mkEnableOption "mousetracker";
          };

          config = mkIf cfg.enable {
            systemd.user.services.mousetracker = {
              Unit.Description = "Foo";
            };
          };
        };
        packages.default = import ./default.nix { inherit pkgs; };
        devShells.default = pkgs.mkShell {
          inputsFrom = [
            self.outputs.packages.${system}.default
          ];
        };
      }
    );
}
