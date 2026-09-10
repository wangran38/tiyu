package main

import (
	"crypto/md5"
	"fmt"
)

func main() {
	h := md5.New()
	h.Write([]byte("123456988cj.comvUSyY6"))
	fmt.Printf("Go MD5:  %x\n", h.Sum(nil))
	h2 := md5.New()
	h2.Write([]byte("123456988cj988cj.comvUSyY6"))
	fmt.Printf("Go MD5b: %x\n", h2.Sum(nil))
	fmt.Printf("Stored:  5c3ba76d17dd649f53271dfac9c09a9f\n")
}
