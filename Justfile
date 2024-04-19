set positional-arguments

# Run dev mode, which includes extra logging
dev *args='': generate
  go run -tags dev . "$@"

# Run a regular build
run *args='': generate
  go run . "$@"

test: generate
  go test -v ./...

mod: 
  go mod tidy

dependencies *args='':
  ./deps.sh "$@"

generate:
  go generate ./...
