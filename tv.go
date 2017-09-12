package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
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

var (
	version_number, build_date string = "unknown", "unknown"
	status int = 0
)

func die() {
	if status > 0 {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
	return
}

func usage() error {
	fmt.Print("tv - showrss handler\n")
	fmt.Printf("build %s, %s\n", version_number, build_date)
	fmt.Print(`
	tv [COMMAND] [ARGUMENTS]

Available commands
	list STRING	print the episodes for show STRING
	fetch STRING	fetch new feed for show STRING
	get STRING ID	copy the magnet link for show STRING, episode ID
	pull		fetch new feed for all shows
	help		display this help text

Interactive mode is started when tv is executed without including the command as a command line parameter. Commands are then entered on the wpa_cli prompt.
`)
	return nil
}

func extract(s, sep string) (string, string) {
	x := strings.Split(s, sep)
	if len(x) < 2 {
		fmt.Fprintf(os.Stderr, "string too short\n")
		return "", ""
	}
	return x[0], x[1]
}

func download(url string, fileName string) error {
	fmt.Fprintf(os.Stdout, "%-10s %s\n", "fetching", url)

	output, err := os.Create(fileName)
	if err != nil {
		status = 1
		return err
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while downloading %v\n%v\n", url, err)
		status = 1
		return err
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	output.Sync()

	fmt.Fprintf(os.Stdout, "%-10s %s %v bytes\n", "fetched", url, n)
	return nil
}

func fetch(showid string) error {
	fd, err := os.Open("subscribed")
	if err != nil {
		status = 1
		return err
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)

	var line string
	for {
		line, err = reader.ReadString('\n')
		if err == io.EOF {
			return fmt.Errorf("download link not found")
		}
		if err != nil {
			fmt.Fprint(os.Stderr, "something's wrong\n")
			status = 1
			return err
		}

		line = strings.TrimSuffix(line, "\n")
		show, url := extract(line, "\t")

		if show == showid {
			err = download(url, "rss/"+showid)
			if err != nil {
				status = 1
				return err
			}
			return nil
		}
	}

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
	for i, eps := range c.Item {
		fmt.Fprintf(pagerStdin, "id\t\t%v\n", i)
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
	files, err := ioutil.ReadDir("rss")
	if err != nil {
		status = 1
		return err
	}

	for _, file := range files {
		err = fetch(file.Name())
		if err != nil {
			status = 1
			return err
		}
	}
	return nil
}

func get(showid string, episodeid string) error {
	epsnr, err := strconv.Atoi(episodeid)
	if err != nil {
		status = 1
		return err
	}

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
	if epsnr >= len(c.Item) {
		status = 1
		return fmt.Errorf("there is no episode %v", epsnr)
	}
	eps := c.Item[epsnr]

	commandToRun := exec.Command("/usr/bin/xclip", "-in")
	commandToRun.Stdout = os.Stdout
	commandToRun.Stderr = os.Stderr
	xclipStdin, err := commandToRun.StdinPipe()
	if err != nil {
		status = 1
		return err
	}

	err = commandToRun.Start()
	if err != nil {
		status = 1
		return err
	}
	fmt.Fprintf(xclipStdin, "%v", eps.Link)

	xclipStdin.Close()

	err = commandToRun.Wait()
	if err != nil {
		status = 1
		return err
	}

	return nil

}

func prompt() error {
	fmt.Fprintf(os.Stdout, "hi!\n")
	for {
		fmt.Fprintf(os.Stdout, "> ")
		reader := bufio.NewReader(os.Stdin)

		line, err := reader.ReadString('\n')
		if err == io.EOF {
			fmt.Fprintf(os.Stdout, "\rgoodbye!\n")
			return nil
		}
		if err != nil {
			status = 1
			return err
		}

		line = strings.TrimSuffix(line, "\n")

		args := strings.Fields(line)

		switch args[0] {
		case "list":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "not enough arguments!\n")
			} else {
				err = list(args[1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
				}
			}
		case "fetch":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "not enough arguments!\n")
				continue
			} else {
				err = fetch(args[1])
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
		case "get":
			if len(args) < 3 {
				fmt.Fprintf(os.Stderr, "not enough arguments!\n")
			} else {
				err = get(args[1], args[2])
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
				}
			}
		case "exit":
			return nil
		default:
			fmt.Fprintf(os.Stderr, "error: unknown command \"%v\"\n", args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		}

	}
	return nil
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
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "not enough arguments!\n")
			status = 1
			return
		} else {
			err = list(os.Args[2])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		}
	case "fetch":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "not enough arguments!\n")
			status = 1
			return
		} else {
			for _, showid := range os.Args[2:] {
				err = fetch(showid)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
				}
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
	case "get":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "not enough arguments!\n")
			status = 1
			return
		}
		err = get(os.Args[2], os.Args[3])
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
