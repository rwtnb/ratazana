{ stdenv, pkgs ? import <nixpkgs-unstable> {} }:

stdenv.mkDeriavation {
  buildInputs = with pkgs; [ 
    gopls
    go_1_15
    go-task
    pkg-config
    xorg.libX11
    xorg.libXcursor
    xorg.libXrandr
    xorg.libXinerama
    xorg.libXi
    xorg.libXext
    xorg.libXxf86vm
    freeglut
    gcc
  ];
}
