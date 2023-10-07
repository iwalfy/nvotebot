{ flake, meta, lib, buildGoApplication, gitignoreSource }:

let
	pname = "nvotebot";
in buildGoApplication {
  inherit pname;
  version = flake.shortRev or (builtins.throw "cannot build with dirty git tree");
  src = gitignoreSource ../.;
  modules = ./gomod2nix.toml;

	doCheck = false;
	subPackages = [ "cmd/${pname}" ];

	CGO_ENABLED = 0;
	ldflags = [ "-w" "-s" "-X main.commitHash=${flake.shortRev}" ];
	tags = [ "netgo" "osusergo" ];

	meta = with lib; {
		description = meta.description;
		homepage = "https://github.com/iwalfy/nvotebot";
		license = licenses.mit;
	};
}
