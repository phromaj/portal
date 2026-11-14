{
  description = "portal: a homelab start page with live service status (Go, stdlib only)";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      forAll = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAll (pkgs: {
        default = pkgs.buildGoModule {
          pname = "portal";
          version = "1";
          src = self;
          vendorHash = null; # standard library only
          env.CGO_ENABLED = 0;
          ldflags = [
            "-s"
            "-w"
          ];
          meta.mainProgram = "portal";
        };
      });

      formatter = forAll (pkgs: pkgs.nixfmt);
    };
}
