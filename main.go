package main

import "github.com/juan-lee/ahabd/cmd"

var (
	version = "unreleased"
)

func main() {
	cmd.Execute(version)
}
