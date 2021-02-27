package main

import (
	"os/user"
	"strings"
)

func ReplacePath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	path = strings.ReplaceAll(path, "~", dir)
	return path
}
