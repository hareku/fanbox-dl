.PHONY: build release

test:
	go test ./...

build:
	go build -o fanbox-dl.out cmd/fanbox-dl/main.go

# make release TAG=vx.x.x
TAG =
release:
	git checkout develop
	git push origin develop
	git checkout main
	git pull origin main
	git merge --ff --no-edit develop
	git push origin main
	git tag ${TAG}
	git push origin tag ${TAG}
	git checkout develop
