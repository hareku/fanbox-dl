.PHONY: build release

build:
	go build -o ./fanboxdl.out ./main.go

TAG =
release:
	git checkout develop;
	git push origin develop;
	git checkout main;
	git merge --no-ff --no-edit develop;
	git push origin main
	git tag ${TAG}
	git push origin tag ${TAG}
	git checkout develop
