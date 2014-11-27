package main

import (
    "bytes"
    "fmt"
    "os"
    "errors"
    "path"
    "regexp"
    "sort"
    "strconv"
    "github.com/codegangsta/cli"
    "path/filepath"
    "io/ioutil"
    "text/template"
    htmltemplate "html/template"
    "github.com/russross/blackfriday"
)

const DEFAULT_TEMPLATE string = "<html><head><title>{{.Filename}}</title></head><body>{{.Body}}</body></html>"
var fileSanitizeRegex = regexp.MustCompile("[-_+.]")

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
    cli.StringFlag{
        Name: "author, a",
        Value: "",
        Usage: "Author name for template: {{.Author}}",
    },
}

// FileAndInfo struct is useful when needing to hold onto
// the original file path when passing around fileinfo
type FileAndInfo struct {
    info os.FileInfo
    path string
}

type ByModTime []FileAndInfo

func (f ByModTime) Len() int {
    return len(f)
}

func (f ByModTime) Swap(i, j int) {
    f[i], f[j] = f[j], f[i]
}

func (f ByModTime) Less(i, j int) bool {
    return f[i].info.ModTime().Before(f[j].info.ModTime())
}

// Page struct is passed into the templates (base and final)
// to assist in generating html files.
type Page struct {
    Body string
    Filename string
    UnixTime string
    Date string
    Author string
}

func main() {
    app := cli.NewApp()
    app.Name = "mdp"
    app.Usage = "Static page generator for markdown source files"
    app.Flags = CLI_FLAGS
    app.Action = run
    app.Run(os.Args)
}

// cleanFilename takes a file name and makes it readable
// so that it can be inserted as standard text into a document
func cleanFilename(filename string) string {
    return fileSanitizeRegex.ReplaceAllString(filename, " ")
}

// handleError takes a string describing an error,
// prints the string and exits with a non-zero int.
func handleError(e error) {
    if e != nil {
        fmt.Println(e.Error())
        os.Exit(1)
    }
}

func run(c *cli.Context) {
    layout := c.String("template")
    if layout != DEFAULT_TEMPLATE {
        templateFile, err := ioutil.ReadFile(layout)
        if err != nil { 
            handleError(errors.New(fmt.Sprintf("Unable to read template file %s: %s", layout, err)))
        }
        layout = string(templateFile)
    }

    matches, _ := filepath.Glob(c.String("input"))
    if len(matches) == 0 {
        handleError(errors.New(fmt.Sprintf("No files found with input \"%s\"", c.String("input"))))
    }

    destDir := c.String("output")

    var finfos []FileAndInfo
    for _, filename := range matches {
        stat, err := os.Stat(filename)
        handleError(err)
        finfos = append(finfos, FileAndInfo{stat, filename})
    }

    sort.Sort(ByModTime(finfos))

    for _, finfo := range finfos {
        err := generateHtmlFile(c, layout, destDir, finfo)
        handleError(err)
    }
}

// generateHtmlFile takes a cli.Context, layout, destination, and source file
// it reads the source markdown file, applies templates, and writes the 
// result to an html file
func generateHtmlFile(c *cli.Context, layout string, destDir string, finfo FileAndInfo) error {
    filename := finfo.path
    time := finfo.info.ModTime()
    extension := filepath.Ext(filename)
    outputName := path.Base(filename[0:len(filename) - len(extension)])

    mdfile, err := ioutil.ReadFile(filename)
    if err != nil {
        return errors.New(fmt.Sprintf("Unable to read %s: %s", filename, err))
    }
    md := blackfriday.MarkdownCommon(mdfile)

    outputFile := path.Join(destDir, outputName + ".html")
    f, err := os.Create(outputFile)
    if err != nil {
        return errors.New(fmt.Sprintf("Unable to create new html file %s: %s", outputFile, err))
    }

    page := Page{
        Body: string(md),
        Filename: cleanFilename(outputName), 
        UnixTime: strconv.FormatInt(time.Unix(), 10),
        Date: time.Format("January 2, 2006"),
        Author: c.String("author"),
    }
    tmp, err := applyTemplate(outputFile, layout, page)

    if err != nil {
        return err
    }

    f.WriteString(tmp)
    f.Close()

    // Set access and modify times so they're useful when
    // programatically generating an index file.
    os.Chtimes(outputFile, time, time)

    // Printing the absolute path to the output file
    // so that it can be potentially piped to other command
    // line tools.
    absPath, _ := filepath.Abs(outputFile)
    fmt.Println(absPath)

    return nil
}

// applyTemplate will generate a template from the layout and apply
// the template to the page twice. The first apply adds
// template objects defined in the layout. The second pass
// adds any template objects defined in the body.
func applyTemplate(templateName string, layout string, page Page) (string, error) {
    btmpl, err := template.New(fmt.Sprintf("%s-base", templateName)).Parse(layout)

    if err != nil {
        return "", errors.New(fmt.Sprintf("Unable to create base template for %s: %s", templateName, err))
    }

    var b bytes.Buffer
    btmpl.Execute(&b, page)

    ftmpl, err := htmltemplate.New(fmt.Sprintf("%s-final", templateName)).Parse(b.String())

    if err != nil {
        return "", errors.New(fmt.Sprintf("Unable to create final template for %s: %s", templateName, err))
    }

    var f bytes.Buffer
    ftmpl.Execute(&f, page)
    return f.String(), nil
}