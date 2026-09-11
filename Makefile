.PHONY: build run test bench clean docker docker-run fmt

build:
go build -ldflags="-s -w" -o bin/nexrouter ./cmd/server

run:
go run cmd/server/main.go

test:
go test ./... -v

bench:
go test ./... -bench=. -benchmem

clean:
rm -rf bin/ coverage.out

docker:
docker build -t nexrouter:latest .

docker-run:
docker run -p 8080:8080 nexrouter:latest

fmt:
go fmt ./...
