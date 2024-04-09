set positional-arguments

# Run dev mode, which includes extra logging
dev *args='': generate
  go run -tags dev . "$@"

# Run a regular build
run *args='': generate
  go run . "$@"

test: generate
  go test -v ./...

update-golden: generate
  go test -v ./tfrender -test.update-golden

mod: 
  go mod tidy

dependencies *args='':
  ./deps.sh "$@"

generate:
  go generate ./...
