package main

import (
	"fmt"
	"os"
)

var status int = 0

func die() {
	if status > 0 {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
	return
}

func usage() error {
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
	return nil
}

func fetch(showid string) error {
	return nil
}

func list(showid string) error {
	return nil
}

func pull() error {
	return nil
}

func prompt() error {
	return fmt.Errorf("nothing prepared yet, status %d\n", status)
}

func main() {
	defer die()

	if len(os.Args) < 2 {
		err = prompt()
		if err != nil {
			status = 1
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		return
	}

	switch os.Args[1] {
	case "list":
		err = list(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	case "fetch":
		for _, showid := range os.Args[2:] {
			err = fetch(showid)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		}
	case "pull":
		err = pull()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	case "help":
		err = usage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command\n")
		err = usage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}

	return
}
