package main

import (
	"coreutils-go/lib"
	"coreutils-go/lib/checksum"
	"crypto/sha256"
)

func main() {
	driver := &checksum.Driver{
		Name:    "SHA224",
		New:     sha256.New224,
		Version: lib.Version,
	}
	driver.Run()
}
