SHELL="/bin/bash"

default: build

dep: 
	@echo 'Please move to vgo and run vgobuild  vgorun'
	@echo 'Run: go get -u golang.org/x/vgo'
	dep ensure

depbuild: dep
	go build cmd/polyman/main.go

run:
	vgo run cmd/polyman/main.go

build:
	vgo build cmd/polyman/main.go


.PHONY: build run depbuild default
