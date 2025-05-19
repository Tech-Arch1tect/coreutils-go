default:
  @just --list

build:
  go build -o build/NotImplemented NotImplemented/main.go
  go build -o build/cp cmd/cp/main.go 
  go build -o build/md5sum cmd/md5sum/main.go
  go build -o build/sha1sum cmd/sha1sum/main.go
  go build -o build/sha224sum cmd/sha224sum/main.go
  go build -o build/sha256sum cmd/sha256sum/main.go
  go build -o build/sha384sum cmd/sha384sum/main.go
  go build -o build/sha512sum cmd/sha512sum/main.go

[working-directory: "testing"]
test:
  docker compose up --build

bt: build test