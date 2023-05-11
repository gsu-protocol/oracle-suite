{ pkgs ? import <nixpkgs> { }, buildGoModule ? pkgs.buildGoModule }:

let
  rev = pkgs.stdenv.mkDerivation {
    name = "rev";
    buildInputs = [ pkgs.git ];
    src = ./.;
    buildPhase = "true";
    installPhase = ''
      echo "$(
        if [ -e .git ]; then
          git rev-parse --short HEAD
        else
          find . -type f -name '*.go' -print0 | sort -z | xargs -0 sha1sum | sha1sum | sed -r 's/[^\da-f]+//g'
        fi
      )" > $out
    '';
  };

  ver = "${pkgs.lib.removeSuffix "\n" (builtins.readFile rev)}";

in buildGoModule {
  pname = "oracle-suite";
  version = builtins.readFile ./version;
  src = ./.;
  vendorSha256 = null;
  subPackages = [ "cmd/..." ];
  postConfigure = "export CGO_ENABLED=0";
  postInstall = "cp ./config.json $out";
}
