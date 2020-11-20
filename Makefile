.PHONY: build merge

build:
	go build -o ./fanboxdl.out ./main.go

merge:
	git checkout main;
	git merge --no-ff develop;
	git push origin main
