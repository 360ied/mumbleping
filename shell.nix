let
  pkgs = import (builtins.fetchTarball {
    name = "nixpkgs-21.11";
    url = "https://github.com/NixOS/nixpkgs/archive/refs/tags/21.11.tar.gz";
    sha256 = "162dywda2dvfj1248afxc45kcrg83appjd0nmdb541hl7rnncf02";
  }) { };
in pkgs.mkShell { nativeBuildInputs = [ pkgs.go_1_17 ]; }
