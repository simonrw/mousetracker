{
  description = "Flake utils demo";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-23.05";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      system = "x86_64-linux";

      pkgs = import nixpkgs {
        inherit system;
      };
    in
    {
      nixosModules.default = { config, lib, ... }:
        with lib;
        let
          cfg = config.mousetracker;
        in
        {
          options.mousetracker = {
            enable = mkEnableOption "mousetracker";
          };

          config = mkIf cfg.enable {
            systemd.user.services.mousetracker = {
              Unit.Description = "Foo";
              Install.WantedBy = [ "graphical-session.target" ];
              Service.ExecStart = "${self.outputs.packages.${system}.default}/bin/mousetracker -flag /dev/input/by-id/usb-Logitech_G203_LIGHTSYNC_Gaming_Mouse_205935534B58-event-mouse -db /tmp/db.db";
            };
          };
        };
      packages.${system}.default = import ./default.nix { inherit pkgs; };
      devShells.${system}.default = pkgs.mkShell {
        inputsFrom = [
          self.outputs.packages.${system}.default
        ];
      };
    };
}
