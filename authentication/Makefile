at = $(shell date +'%FT%T%:z')

#### cmd
build:
	echo ">>> ${at}"
	bash scripts/go_build.sh authentication
	ls -lh target/authentication

run:
	echo ">>> ${at}"
	bash scripts/go_build.sh authentication
	./target/authentication
