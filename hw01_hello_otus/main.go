// pckage main is main package
package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse" //nolint:depguard
)

func main() {
	fmt.Println(reverse.String("Hello, OTUS!"))
}
