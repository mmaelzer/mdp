package main

import (
    "os"
    "path"
    "github.com/codegangsta/cli"
    "path/filepath"
    "io/ioutil"
    "text/template"
    "github.com/russross/blackfriday"
)

var DEFAULT_TEMPLATE = "<html><head><title>{{.Title}}</title></head><body>{{.Body}}</body></html>"

var CLI_FLAGS = []cli.Flag {
    cli.StringFlag{
        Name: "input, i",
        Value: "*.md",
        Usage: "Markdown files to process",
    },
    cli.StringFlag{
        Name: "template, t",
        Value: DEFAULT_TEMPLATE,
        Usage: "Template file to use for generating HTML",
    },
    cli.StringFlag{
        Name: "output, o",
        Value: "./",
        Usage: "Location to write HTML files to",
    },
}

type Page struct {
    Body string
    Title string
}

func main() {
    app := cli.NewApp()
    app.Name = "markdown-for-what"
    app.Usage = "Static site generator for markdown source files"
    app.Flags = CLI_FLAGS
    app.Action = func(c *cli.Context) {     
        layout := c.String("template")
        if layout != DEFAULT_TEMPLATE {
            templateFile, err := ioutil.ReadFile(layout)
            if err != nil { panic(err) }
            layout = string(templateFile)
        }
        matches, _ := filepath.Glob(c.String("input"))
        destDir := c.String("output")
        for _, filename := range matches {
            var extension = filepath.Ext(filename)
            var outputName = path.Base(filename[0:len(filename) - len(extension)])

            mdfile, err := ioutil.ReadFile(filename)
            if err != nil { panic(err) }

            md := blackfriday.MarkdownCommon(mdfile)
            page := Page{string(md), outputName}

            tmpl, err := template.New(outputName).Parse(layout)

            if err != nil { panic(err) }

            f, err := os.Create(path.Join(destDir, outputName + ".html"))

            if err != nil { panic(err) }

            tmpl.Execute(f, page)

            f.Close()
        }
    }
    app.Run(os.Args)
}