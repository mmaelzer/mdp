Markdown Pages
=================

A CLI that takes markdown files as input and outputs HTML pages.

Example
-------
```bash
$ git clone git@github.com:mmaelzer/mdp
$ cd mdp
$ go build mdp.md
$ mdp -i "~/mysite/src/*.md" -o ~/mysite/html/ -t ~/mysite/src/layout.html
/home/myuser/mysite/html/post1.html
/home/myuser/mysite/html/post2.html
/home/myuser/mysite/html/post3.html
```


Template Objects
----------------
### {{.Body}}
The html generated from the source markdown file.

### {{.Filename}}
The filename of the markdown file with the extension removed and characters `-`, `+`, `_`, and `.` replaced by whitespace.

### {{.UnixTime}}
A unix timestamp based on the markdown file's last modify time.

### {{.Date}}
A date string in the format `January 2, 2006`. The time is based on the markdown file's last modify time.

### {{.Author}}
The string provided by the command line argument `-a` or `--author`.


Usage
-------
```
NAME:
   mdp - Static page generator for markdown source files

USAGE:
   mdp [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --input, -i '*.md'
                Markdown files to process
   --template, -t '<html><head><title>{{.Filename}}</title></head><body>{{.Body}}</
body></html>'   Template file to use for generating HTML
   --output, -o './'
                Location to write HTML files to
   --author, -a ''
                Author name for template: {{.Author}}
   --help, -h
                show help
   --version, -v
                print the version
```