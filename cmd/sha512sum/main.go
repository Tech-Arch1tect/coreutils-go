package main

import (
	"coreutils-go/lib"
	"coreutils-go/lib/checksum"
	"crypto/sha512"
)

func main() {
	driver := &checksum.Driver{
		Name:    "SHA512",
		New:     sha512.New,
		Version: lib.Version,
	}
	driver.Run()
}
