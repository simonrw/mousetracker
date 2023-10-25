{ pkgs ? import <nixpkgs> {} }:
with pkgs;
buildGoModule {
  pname = "evdevmonitor";
  version = "unstable";

  src = ./.;

  # dependencies are vendored, so no vendor hash required
  vendorHash = null;
}
