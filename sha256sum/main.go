package main

import (
	"coreutils-go/lib"
	"coreutils-go/lib/checksum"
	"crypto/sha256"
)

func main() {
	driver := &checksum.Driver{
		Name:    "SHA256",
		New:     sha256.New,
		Version: lib.Version,
	}
	driver.Run()
}
