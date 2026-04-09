{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
  };

  outputs = { self, nixpkgs, ... }: 
  let
    pkgs = import nixpkgs { 
      system = "x86_64-linux"; 
      config.allowUnfreePredicate = pkg: builtins.elem (nixpkgs.lib.getName pkg) [
	# Put unfree packages here
      ];
    };
  in 
  {
    devShell.x86_64-linux = pkgs.mkShell {
      packages = with pkgs; [
	go
      ];
    };
  };
}
