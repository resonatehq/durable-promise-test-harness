.PHONY: deps
deps: 
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest3

.PHONY: gen
gen: 
	oapi-codegen -generate types,client -package openapi ../durable-promise/spec/durable-promise.yaml > pkg/simulator/openapi/openapi.go