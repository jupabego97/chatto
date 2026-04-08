package main

import "hmans.de/chatto/cmd"

func main() {
	cmd.SetVersion(Version)
	cmd.Execute()
}
