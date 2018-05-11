package main

import (
	"fmt"
)

func printInfo(item string) {
	fmt.Printf("\033[0;32m%s\033[0m\n", item)
}

func printWarn(item string) {
	fmt.Printf("\033[0;91m%s\033[0m\n", item)
}

func printPair(key string, value string) {
	fmt.Printf("\033[0;34m%s: \033[0;35m%s\033[0m\n", key, value)
}
