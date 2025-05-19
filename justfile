default:
  @just --list

build:
  go build -o build/NotImplemented NotImplemented/main.go
  go build -o build/cp cp/main.go 
  go build -o build/md5sum md5sum/main.go
  go build -o build/sha1sum sha1sum/main.go

[working-directory: "testing"]
test:
  docker compose up --build

bt: build test