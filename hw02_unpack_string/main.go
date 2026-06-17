// Package main package need to demonstrate work of unpacking
package main

import (
	"fmt"

	"github.com/shidemere/2026-05-golang-professional/hw02_unpack_string/unpack"
)

func main() {
	out, err := unpack.Unpack("🙃0")
	if err != nil {
		fmt.Printf("Произошла ошибка: %v", err)
		return
	}
	fmt.Println(out)
}
