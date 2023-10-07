rec {
	description = "Telegram bot for Neural OpenNet (https://t.me/neuro_opennet) post voting system (https://nvote.lebedinets.ru/)";

	inputs = {
		nixpkgs.url = "nixpkgs/nixos-unstable";
		flake-parts.url = "github:hercules-ci/flake-parts";
		systems.url = "github:nix-systems/default";

		gitignore = {
			url = "github:hercules-ci/gitignore.nix";
			inputs.nixpkgs.follows = "nixpkgs";
		};

		gomod2nix = {
			url = "github:nix-community/gomod2nix";
			inputs.nixpkgs.follows = "nixpkgs";
		};
	};

	outputs = inputs @ { self, nixpkgs, flake-parts, gitignore, gomod2nix, ... }:
		flake-parts.lib.mkFlake { inherit inputs; } {
			systems = import inputs.systems;

			perSystem = { pkgs, system, ... }: let
				inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
				inherit (gitignore.lib) gitignoreSource;
			in {
				packages = rec {
					nvotebot = pkgs.callPackage ./nix/nvotebot.nix {
						inherit buildGoApplication gitignoreSource;
						flake = self;
						meta = { inherit description; };
					};

					default = nvotebot;
				};
			};
		};
}
