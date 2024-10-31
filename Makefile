.PHONY: build release

test:
	go test ./...

build:
	goreleaser build --single-target --skip=validate --clean

# make tag TAG=vx.x.x
tag:
	git checkout develop
	git push origin develop
	git checkout main
	git pull origin main
	git merge --ff --no-edit develop
	git push origin main
	git tag ${TAG}
	git push origin tag ${TAG}
	git checkout develop

release:
	goreleaser release --snapshot --clean
	goreleaser check
	goreleaser release
