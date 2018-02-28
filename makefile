SHELL="/bin/bash"

default: build

run:
	vgo run cmd/polyman/main.go

build:
	vgo build -o polyman cmd/polyman/main.go


.PHONY: build run default
