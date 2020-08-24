package main

import (
	"fmt"

	hc "github.com/sirmackk/healthchecker"
)

// break out into non-main package
// introduce interfaces
// add tests - unit and integration



func main() {
	checker := hc.NewHTTPChecker()
	result := checker.SimpleHTTPCheck("http://mattscodecave.com")
	fmt.Println(result)
}
