sources := $(shell find "$(PWD)" -name '*.go')

.PHONY: all get test generate

all: generate

# Get dependencies
get:
	go get -u github.com/tinylib/msgp

# Test
test: generate
	go test ./...

# Generate
generate: ./models/z_gen_content.go ./models/z_gen_meta.go ./models/z_gen_proto.go 

./models/z_gen_content.go: ./models/content.go
	msgp -tests=false -file $< -o $@

./models/z_gen_meta.go: ./models/meta.go
	msgp -tests=false -file $< -o $@

./models/z_gen_proto.go : ./models/proto.go
	msgp -tests=false -file $< -o $@
