package cmd

import (
	"archive/zip"
	"context"
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
	"github.com/urfave/cli"
)

const owner = "hakimel"
const repo = "reveal.js"
const templateIndex = `<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
		<link rel="stylesheet" href="css/reveal.css">
		<link rel="stylesheet" href="css/theme/{{.Theme}}.css" id="theme">
		<link rel="stylesheet" href="lib/css/zenburn.css">
		<script src="livereload.js"></script>
	</head>
	<body>
		<div class="reveal">
			<div class="slides">
				<section data-markdown="{{.Slide}}"
					data-separator="^\r?\n\r?\n\r?\n"
					data-separator-vertical="^\r?\n---\r?\n$"
				>
				</section>
			</div>
		</div>
		<script src="lib/js/head.min.js"></script>
		<script src="js/reveal.js"></script>
		<script>
			Reveal.initialize({
				controls: true,
				progress: true,
				history: true,
				center: true,
				theme: Reveal.getQueryHash().theme, // available themes are in /css/theme
				transition: Reveal.getQueryHash().transition || 'default', // default/cube/page/concave/zoom/linear/fade/none
				// Optional libraries used to extend on reveal.js
				dependencies: [
					{ src: 'lib/js/classList.js', condition: function() { return !document.body.classList; } },
					{ src: 'plugin/markdown/marked.js', condition: function() { return !!document.querySelector( '[data-markdown]' ); } },
					{ src: 'plugin/markdown/markdown.js', condition: function() { return !!document.querySelector( '[data-markdown]' ); } },
					{ src: 'plugin/highlight/highlight.js', async: true, callback: function() { hljs.initHighlightingOnLoad(); } },
					{ src: 'plugin/zoom-js/zoom.js', async: true, condition: function() { return !!document.body.classList; } },
					{ src: 'plugin/notes/notes.js', async: true, condition: function() { return !!document.body.classList; } }
				]
			});
		</script>
	</body>
</html>`

var initCmd = cli.Command{
	Name:    "init",
	Aliases: []string{"i"},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "title",
			Value: "no title",
		},
		cli.StringFlag{
			Name:  "theme",
			Value: "black",
		},
		cli.StringFlag{
			Name:  "slide",
			Value: "slide.md",
		},
	},
	Action: initAction,
}

type HtmlValue struct {
	Title string
	Theme string
	Slide string
}

func initAction(c *cli.Context) error {
	client := github.NewClient(nil)
	ctx := context.Background()

	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return err
	}
	zip, err := download(release.GetZipballURL())
	if err != nil {
		return err
	}
	defer func() {
		os.Remove(zip)
	}()
	base, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	root := filepath.Join(base, c.GlobalString("root"))
	if err = extract(zip, root); err != nil {
		return err
	}
	value := HtmlValue{
		Theme: c.String("theme"),
		Title: c.String("title"),
		Slide: c.String("slide"),
	}
	tmpl := template.Must(template.New("index").Parse(templateIndex))
	html, err := os.Create(filepath.Join(root, "index.html"))
	if err != nil {
		return err
	}
	defer html.Close()
	if err = tmpl.Execute(html, value); err != nil {
		return err
	}
	md, err := os.Create(filepath.Join(root, c.String("slide")))
	if err != nil {
		return err
	}
	defer md.Close()
	if _, err = md.Write([]byte(c.String("title"))); err != nil {
		return err
	}
	return nil
}

func download(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	t, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	io.Copy(t, resp.Body)

	return t.Name(), nil
}

func extract(path, out string) error {
	rc, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer rc.Close()
	root, err := func() (string, error) {
		if len(rc.File) < 1 {
			return "", errors.New("no entry.")
		}
		return rc.File[0].Name, nil
	}()
	if err != nil {
		return err
	}
	for i, f := range rc.File {
		if i == 0 {
			continue
		}
		path := relative(f.Name, root)
		if ignore(path, root) {
			continue
		}
		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(filepath.Join(out, path), 0700); err != nil {
				return err
			}
			continue
		}
		r, err := f.Open()
		if err != nil {
			return err
		}
		w, err := os.Create(filepath.Join(out, path))
		if err != nil {
			return err
		}
		if _, err = io.Copy(w, r); err != nil {
			return err
		}
	}
	return nil
}

func relative(path, root string) string {
	return strings.Replace(path, root, "", -1)
}

func ignore(entry, root string) bool {
	entry = strings.Replace(entry, root, "", -1)
	return !strings.HasPrefix(entry, "js") &&
		!strings.HasPrefix(entry, "css") &&
		!strings.HasPrefix(entry, "lib") &&
		!strings.HasPrefix(entry, "plugin")
}

func init() {
	addCommand(initCmd)
}
