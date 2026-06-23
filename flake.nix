{
  description = "Pixiv FANBOX downloader";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
      rev = self.shortRev or self.dirtyShortRev or "dev";
    in
    {
      packages = forAllSystems (system:
        let pkgs = nixpkgs.legacyPackages.${system}; in
        {
          default = pkgs.buildGoModule {
            pname = "fanbox-dl";
            version = rev;

            src = ./.;
            vendorHash = "sha256-o9/e8IbWN7PzOG60B2IBwVYIWM1qTeYC4EdubdG179s=";

            subPackages = [ "cmd/fanbox-dl" ];

            env.CGO_ENABLED = 0;

            ldflags = [
              "-s"
              "-w"
              "-X=main.version=${rev}"
              "-X=main.commit=${self.rev or self.dirtyRev or "none"}"
            ];

            meta = with pkgs.lib; {
              description = "Pixiv FANBOX downloader";
              homepage = "https://github.com/hareku/fanbox-dl";
              license = licenses.mit;
              mainProgram = "fanbox-dl";
            };
          };
        });
    };
}
