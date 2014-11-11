package main

import (
    "fmt"
    "os"
    "path"
    "regexp"
    "github.com/codegangsta/cli"
    "path/filepath"
    "io/ioutil"
    "text/template"
    "github.com/russross/blackfriday"
)

var DEFAULT_TEMPLATE = "<html><head><title>{{.Filename}}</title></head><body>{{.Body}}</body></html>"

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

var fileSanitizeRegex = regexp.MustCompile("[-_+.]")

type Page struct {
    Body string
    Filename string
}

func main() {
    app := cli.NewApp()
    app.Name = "markdown-for-what"
    app.Usage = "Static site generator for markdown source files"
    app.Flags = CLI_FLAGS
    app.Action = run
    app.Run(os.Args)
}

func cleanFilename(filename string) string {
    return fileSanitizeRegex.ReplaceAllString(filename, " ")
}

func run(c *cli.Context) {
    layout := c.String("template")
    if layout != DEFAULT_TEMPLATE {
        templateFile, err := ioutil.ReadFile(layout)
        if err != nil { 
            fmt.Printf("Unable to read template file %s: %s\n", layout, err)
            os.Exit(1)
        }
        layout = string(templateFile)
    }

    matches, _ := filepath.Glob(c.String("input"))
    if len(matches) == 0 {
        fmt.Printf("No files found with input \"%s\"\n", c.String("input"))
        os.Exit(1)
    }

    destDir := c.String("output")

    for _, filename := range matches {
        extension := filepath.Ext(filename)
        outputName := path.Base(filename[0:len(filename) - len(extension)])

        mdfile, err := ioutil.ReadFile(filename)
        if err != nil { 
            fmt.Printf("Unable to read %s: %s\n", filename, err)
            os.Exit(1)
        }

        md := blackfriday.MarkdownCommon(mdfile)
        page := Page{string(md), cleanFilename(outputName)}

        tmpl, err := template.New(outputName).Parse(layout)

        if err != nil { 
            fmt.Printf("Unable to create template for %s: %s\n", filename, err)
            os.Exit(1)
        }

        outputFilename := outputName + ".html"
        f, err := os.Create(path.Join(destDir, outputFilename))
        if err != nil { 
            fmt.Printf("Unable to create new html file %s: %s\n", outputFilename, err)
            os.Exit(1)
        }

        tmpl.Execute(f, page)

        f.Close()
    }
}