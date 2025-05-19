default:
  @just --list

build:
  go build -o build/NotImplemented NotImplemented/main.go
  go build -o build/cp cp/main.go 
  go build -o build/md5sum md5sum/main.go
  go build -o build/sha1sum sha1sum/main.go
  go build -o build/sha224sum sha224sum/main.go
  go build -o build/sha256sum sha256sum/main.go
  go build -o build/sha384sum sha384sum/main.go
  go build -o build/sha512sum sha512sum/main.go

[working-directory: "testing"]
test:
  docker compose up --build

bt: build test