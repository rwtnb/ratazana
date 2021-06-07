{ pkgs ? import <nixpkgs-unstable> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [ 
    gopls
    go_1_16
    go-task
    pkg-config
    xorg.libX11
    xorg.libXcursor
    xorg.libXrandr
    xorg.libXinerama
    xorg.libXi
    xorg.libXext
    xorg.libXxf86vm
    libGL
    gcc
  ];
}
