gitBranch = $(shell git rev-parse --abbrev-ref HEAD)
gitTime = $(shell date +'%FT%T%:z')

#### cmd
build:
	echo ">>> git branch: ${gitBranch}, git time: ${gitTime}"
	bash scripts/go_build.sh authentication
	ls -lh target/authentication

run:
	bash scripts/go_build.sh authentication
	./target/authentication
