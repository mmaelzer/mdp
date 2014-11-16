package main

import (
    "fmt"
    "os"
    "errors"
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

func handleError(errorStr string) {
    fmt.Println(errorStr)
    os.Exit(1)
}

func run(c *cli.Context) {
    layout := c.String("template")
    if layout != DEFAULT_TEMPLATE {
        templateFile, err := ioutil.ReadFile(layout)
        if err != nil { 
            handleError(fmt.Sprintf("Unable to read template file %s: %s", layout, err))
        }
        layout = string(templateFile)
    }

    matches, _ := filepath.Glob(c.String("input"))
    if len(matches) == 0 {
        handleError(fmt.Sprintf("No files found with input \"%s\"", c.String("input")))
    }

    destDir := c.String("output")

    for _, filename := range matches {
        err := generateHtmlFile(c, layout, destDir, filename)
        if err != nil {
            handleError(err.Error())
        }
    }
}

func generateHtmlFile(c *cli.Context, layout string, destDir string, filename string) error {
    extension := filepath.Ext(filename)
    outputName := path.Base(filename[0:len(filename) - len(extension)])
    srcFstat, err := os.Stat(filename)

    if err != nil {
        return errors.New(fmt.Sprintf("Unable to get FileInfo for %s", filename))
    }

    mdfile, err := ioutil.ReadFile(filename)
    if err != nil {
        return errors.New(fmt.Sprintf("Unable to read %s: %s", filename, err))
    }

    md := blackfriday.MarkdownCommon(mdfile)
    page := Page{string(md), cleanFilename(outputName)}

    tmpl, err := template.New(outputName).Parse(layout)

    if err != nil {
        return errors.New(fmt.Sprintf("Unable to create template for %s: %s", filename, err))
    }

    outputFile := path.Join(destDir, outputName + ".html")
    f, err := os.Create(outputFile)
    if err != nil {
        return errors.New(fmt.Sprintf("Unable to create new html file %s: %s", outputFile, err))
    }

    tmpl.Execute(f, page)

    f.Close()

    // Set access and modify times so they're useful when
    // programatically generating an index file.
    os.Chtimes(outputFile, srcFstat.ModTime(), srcFstat.ModTime())

    // Printing the absolute path to the output file
    // so that it can be potentially piped to other command
    // line tools.
    absPath, _ := filepath.Abs(outputFile)
    fmt.Println(absPath)

    return nil
}