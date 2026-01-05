default: build

test:
	go test ./...

build:
	go build -o tflint-ruleset-tmn

install: build
	mkdir -p ~/.tflint.d/plugins
	mv ./tflint-ruleset-tmn ~/.tflint.d/plugins
