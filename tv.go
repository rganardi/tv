package main

import (
	"fmt"
	"os"
)

var status int = 0

func die(status int) {
	if status > 0 {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
	return
}

func usage() {
	fmt.Print(`
tv - showrss handler

	tv [COMMAND] [ARGUMENTS]

Available commands
	list STRING	print the episodes for show STRING
	fetch STRING	fetch new feed for show STRING
	pull		fetch new feed for all shows
	help		display this help text

Interactive mode is started when tv is executed without including the command as a command line parameter. Commands are then entered on the wpa_cli prompt.
`)
	return
}

func fetch(showid string) {
}

func list(showid string) {
}

func pull() {
}

func prompt() {
}

func main() {
	defer die(status)

	if len(os.Args) < 2 {
		prompt()
		return
	}

	switch os.Args[1] {
	case "list":
		list(os.Args[2])
	case "fetch":
		fetch(os.Args[2])
	case "pull":
		pull()
	case "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command\n")
		usage()
	}

	return
}
