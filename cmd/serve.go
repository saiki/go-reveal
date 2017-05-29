package cmd

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/omeid/go-livereload"
	"github.com/pkg/browser"
	"github.com/urfave/cli"
)

var serveCmd = cli.Command{
	Name:    "serve",
	Aliases: []string{"s"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "address, a",
			Value: ":8080",
		},
		cli.BoolFlag{
			Name: "watch, w",
		},
	},
	Action: serveAction,
}

func serveAction(c *cli.Context) error {
	root, err := filepath.Abs(c.GlobalString("root"))
	if err != nil {
		return err
	}
	addr := c.String("address")
	log.SetOutput(browser.Stdout)
	var url string
	if strings.HasPrefix(addr, ":") {
		url = "http://localhost" + addr
	} else {
		url = "http://" + addr
	}
	browser.OpenURL(url)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(root)))
	mux.HandleFunc("/livereload.js", livereload.LivereloadScript)
	server := livereload.New("slide server")
	defer server.Close()
	mux.Handle("/livereload", server)
	lrc, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer lrc.Close()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					server.Reload(url, false)
				}
			case err := <-watcher.Errors:
				log.Println(err)
			}
		}
	}()
	if err = watcher.Add(root); err != nil {
		return err
	}

	return http.Serve(lrc, mux)
}

func init() {
	addCommand(serveCmd)
}
