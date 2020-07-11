BIN=bin
BIN_NAME=bom-maker

all: build

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: build
build $(BIN)/$(BIN_NAME): $(BIN)
	env CGO_ENABLED=0 go build -o $(BIN)/$(BIN_NAME)

$(BIN):
	mkdir -p $(BIN)

.PHONY: clean
clean:
	rm -rf $(BIN)

.PHONY: test
test:
	go test `go list ./... | grep -v /vendor/`
