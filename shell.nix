with import <nixpkgs> { };

stdenv.mkDerivation rec {
  name = "moggio";
  buildInputs = with pkgs; [ go_1_16 libpulseaudio pkg-config ];
  shellHook = ''
    export PS1='nix@\w$(__git_ps1 " (%s)")\$ '
  '';
}
