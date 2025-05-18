default:
  @just --list

build:
  go build -o build/NotImplemented NotImplemented/main.go
  go build -o build/cp cp/main.go 
