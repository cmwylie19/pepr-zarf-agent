# Makefile for building the Pepr Zarf Agent and the Transformer Service

SHELL=bash
DOCKER_USERNAME=cmwylie19
TAG=0.0.1

include wasm-transform/Makefile

.PHONY: build/pepr-zarf-agent
build/pepr-zarf-agent:
	@echo "Building Pepr Zarf Agent"
	@echo "Create kind cluster"
	@kind create cluster --name=pepr-zarf-agent
	@pepr build



.PHONY: build/wasm-transform
build/wasm-transform: 
	@cd wasm-transform
	@$(MAKE) -C wasm-transform build/wasm-transformer


.PHONY: deploy/dev
deploy/dev:
	@echo "Deploying to Dev"
	@kubectl create -k transformer/k8s/overlays/dev 
	@sleep 5
	@kubectl wait --for=condition=Ready pod -l app=transformer --timeout=60s -n pepr-system
	@kubectl wait --for=condition=Ready pod -l run=debugger --timeout=60s -n pepr-system


.PHONY: clean
clean:
	@echo "Removing cluster"
	@kind delete cluster --name=pepr-zarf-agent


all: build/wasm-transform build/pepr-zarf-agent
