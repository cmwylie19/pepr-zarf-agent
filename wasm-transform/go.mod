module github.com/cmwylie19/pepr-zarf-agent/wasm-transform

go 1.20

replace github.com/defenseunicorns/zarf v0.28.1 => github.com/cmwylie19/zarf v0.28.3

require github.com/defenseunicorns/zarf v0.28.1

require (
	github.com/distribution/distribution v2.8.2+incompatible // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
)
