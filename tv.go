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
	subscribed string = "subscribed"
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
	response, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error while downloading %v\n", url)
		status = 1
		return err
	}
	defer response.Body.Close()

	if (response.StatusCode != http.StatusOK) {
		status = 1
		return fmt.Errorf("HTTP status code : %v", response.StatusCode)
	}

	//check if the received file is a showrss rss
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		status = 1
		return err
	}

	q := query{}
	d := xml.NewDecoder(bytes.NewReader(body))

	err = d.Decode(&q)
	if err != nil {
		status = 1
		return fmt.Errorf("downloaded file is not an rss feed")
	}

	err = ioutil.WriteFile(fileName, body, 0644)
	if err != nil {
		status = 1
		return err
	}

	return nil
}

func fetch(showid string) error {
	fd, err := os.Open(subscribed)
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

	for i, eps := range c.Item {
		fmt.Fprintf(os.Stdout, "id\t\t%v\n", i)
		fmt.Fprintf(os.Stdout, "title\t\t%v\n", eps.Title)
		fmt.Fprintf(os.Stdout, "date\t\t%v\n", eps.PubDate)
		fmt.Fprintf(os.Stdout, "link\t\t%v\n", eps.Link)
		fmt.Fprintf(os.Stdout, "\n")
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
		fmt.Fprintf(os.Stdout, "%s %-50s\r", "fetching", file.Name())
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

	fmt.Fprintf(os.Stdout, "copied link for \"%v\" into clipboard\n", eps.Title)

	return nil

}

func main() {
	defer die()

	err := os.Chdir(os.Getenv("HOME") + "/tv/")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}

	if len(os.Args) < 2 {
		status = 1
		fmt.Fprintf(os.Stderr, "not enough arguments!\n")
		_ = usage()
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
