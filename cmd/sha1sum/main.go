package main

import (
	"coreutils-go/lib"
	"coreutils-go/lib/checksum"
	"crypto/sha1"
)

func main() {
	driver := &checksum.Driver{
		Name:    "SHA1",
		New:     sha1.New,
		Version: lib.Version,
	}
	driver.Run()
}
