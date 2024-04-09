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
  go test -v ./... -test.update-golden

mod: 
  go mod tidy

dependencies *args='':
  ./deps.sh "$@"

generate:
  just dependencies github.com/sqlc-dev/sqlc/cmd/sqlc
  go generate ./...
