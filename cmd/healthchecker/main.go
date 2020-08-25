package main

import (
	"fmt"

	hc "github.com/sirmackk/healthchecker"
)

// introduce interfaces



func main() {
	checker := hc.NewHTTPChecker()
	result := checker.SimpleHTTPCheck("http://mattscodecave.com")
	fmt.Println(result)
}
