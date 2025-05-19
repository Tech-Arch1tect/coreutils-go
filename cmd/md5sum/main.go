package main

import (
	"coreutils-go/lib"
	"coreutils-go/lib/checksum"
	"crypto/md5"
)

func main() {
	driver := &checksum.Driver{
		Name:    "MD5",
		New:     md5.New,
		Version: lib.Version,
	}
	driver.Run()
}
