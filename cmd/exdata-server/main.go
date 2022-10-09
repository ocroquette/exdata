package main

import "github.com/ocroquette/exdata/internal/exdata"

func main() {
	server := new(exdata.Server)
	server.Start("localhost:8080", "repo")
}
