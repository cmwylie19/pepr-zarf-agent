# Makefile for building WASM Transform

SHELL=bash
DOCKER_USERNAME=cmwylie19
TAG=0.0.1

build/wasm-transformer:
	@echo "Building WASM Transform"
	@GOOS=js GOARCH=wasm go build -o main.wasm
	@mv main.wasm ../wasm-integration-test/assets/
