{ pkgs ? import <nixpkgs> {} }:
with pkgs;
buildGoModule {
  pname = "mousetracker";
  version = "unstable";

  src = ./.;

  # dependencies are vendored, so no vendor hash required
  vendorHash = null;
}
