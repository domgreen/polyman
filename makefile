SHELL="/bin/bash"

run:
	vgo run cmd/polyman/main.go

build:
	vgo build cmd/polyman/main.go

.PHONY: build run
