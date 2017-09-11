package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type query struct {
	Channel channel `xml:"channel"`
}

type channel struct {
	Title string    `xml:"title"`
	Item  []episode `xml:"item"`
}

type episode struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
}

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
	fd, err := ioutil.ReadFile("rss/" + showid)
	if err != nil {
		status = 1
		return err
	}

	q := query{}
	d := xml.NewDecoder(bytes.NewReader(fd))

	err = d.Decode(&q)
	if err != nil {
		status = 1
		return err
	}

	c := q.Channel

	env := os.Environ()
	if err != nil {
		fmt.Fprint(os.Stderr, "error getting environment variables\n")
		status = 1
		return err
	}

	pager := "/usr/bin/less"
	for _, variable := range env {
		if strings.HasPrefix(variable, "PAGER") {
			pager = strings.TrimPrefix(variable, "PAGER=")
		}
	}

	commandToRun := exec.Command(pager)
	commandToRun.Stdout = os.Stdout
	pagerStdin, err := commandToRun.StdinPipe()
	if err != nil {
		status = 1
		return err
	}

	err = commandToRun.Start()
	if err != nil {
		status = 1
		return err
	}
	for _, eps := range c.Item {
		fmt.Fprintf(pagerStdin, "title\t\t%v\n", eps.Title)
		fmt.Fprintf(pagerStdin, "date\t\t%v\n", eps.PubDate)
		fmt.Fprintf(pagerStdin, "link\t\t%v\n", eps.Link)
		fmt.Fprintf(pagerStdin, "\n")
	}

	pagerStdin.Close()

	err = commandToRun.Wait()
	if err != nil {
		status = 1
		return err
	}

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

	err := os.Chdir(os.Getenv("HOME") + "/tv/")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}

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
