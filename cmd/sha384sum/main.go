package main

import (
	"coreutils-go/lib"
	"coreutils-go/lib/checksum"
	"crypto/sha512"
)

func main() {
	driver := &checksum.Driver{
		Name:    "SHA384",
		New:     sha512.New384,
		Version: lib.Version,
	}
	driver.Run()
}
