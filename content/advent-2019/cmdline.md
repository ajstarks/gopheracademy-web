+++
title = "Writing Friendly Command Line Applications"
date = "2019-12-02T00:00:00+00:00"
series = ["Advent 2019"]
author = ["Miki Tebeka"]
linktitle = "flag"
+++

Let me tell you a story...

In 1986 [Knuth](https://en.wikipedia.org/wiki/Donald_Knuth) wrote a program to
demonstrate [literate
programming](https://en.wikipedia.org/wiki/Literate_programming).

The task was to read a file of text, determine the n most frequently used
words, and print out a sorted list of those words along with their frequencies.
Knuth wrote a beautiful 10 page monolithic program.

Doug Mcllory read this and said 
`tr -cs A-Za-z '\n' | tr A-Z a-z | sort | uniq -c | sort -rn | sed ${1}q`

It's 2019, why am I telling you a story that happened 33 years ago? (Probably
before some of you were born). The computation landscape has changed a lot...
or has it?

The [Lindy effect](https://en.wikipedia.org/wiki/Lindy_effect) is a concept
that the future life expectancy of some non-perishable things like a technology
or an idea is proportional to their current age. TL;DR - old technologies are
here to stay.

If you don't believe me, see:

- [oh-my-zsh](https://github.com/ohmyzsh/ohmyzsh) having close to 100,000 stars on GitHub
- [Data Science at the Command Line](https://www.datascienceatthecommandline.com/) book
- [Command-line Tools can be 235x Faster than your Hadoop Cluster](https://adamdrake.com/command-line-tools-can-be-235x-faster-than-your-hadoop-cluster.html)
- ...

Now that you are convinced, let's talk on how to make your Go programs command
line friendly.

## Design

When writing command line application, try to adhere to the [basics of Unix
philosophy](http://www.catb.org/esr/writings/taoup/html/ch01s06.html)

- Rule of Modularity: Write simple parts connected by clean interfaces.
- Rule of Composition: Design programs to be connected with other programs.
- Rule of Silence: When a program has nothing surprising to say, it should say
  nothing.

These rules allow you to write small program that do one thing.

- A user asks for support of reading data from REST API? Have them pipe a
  `curl` command output to your program
- A user wants only top n results? Have them pipe your program output through
  `head
- A user wants only the second column of data? Since you write tab seperated
  output, they can pipe your output via `cut` or `awk`

If you don't follow these and let your command line interface grow organically,
you might end up in the following situation

[![](https://imgs.xkcd.com/comics/tar.png)](https://xkcd.com/1168/)


## Help

Let's assume your team have a `nuke-db` utility. You forgot how to invoke it
and you do:

```
$ ./nuke-db --help
database nuked
```

Ouch!

Using the [flag](https://golang.org/pkg/flag/), you can add support for `--help` in 2 extra lines of code

```go
package main

import (
	"flag" // extra line 1
	"fmt"
)

func main() {
	flag.Parse() // extra line 2
	fmt.Println("database nuked")
}
```

Now your program behaves

```
$ ./nuke-db --help
Usage of ./nuke-db:
$ ./nuke-db
database nuked
```

If you'd like to provide more help, use `flag.Usage`

```go
package main

import (
	"flag"
	"fmt"
	"os"
)

var usage = `usage: %s [DATABASE]

Delete all data and tables from DATABASE.
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	fmt.Println("database nuked")
}
```

And now
```
$ ./nuke-db --help
usage: ./nuke-db [DATABASE]

Delete all data and tables from DATABASE.
```

## Structured Output

Plain text is the universal interface. However, when the output becomes
complex, it might be easier for machines to deal with formatted output. One of
the most common format is of course JSON.

A good way to do it is not to print using `fmt.Printf` but use your own
printing function which can be either text or JSON. Let's see an example:

```go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var jsonOut bool
	flag.BoolVar(&jsonOut, "json", false, "output in JSON format")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("error: wrong number of arguments")
	}

	write := writeText
	if jsonOut {
		write = writeJSON
	}

	fi, err := os.Stat(flag.Arg(0))
	if err != nil {
		log.Fatalf("error: %s\n", err)
	}

	m := map[string]interface{}{
		"size":     fi.Size(),
		"dir":      fi.IsDir(),
		"modified": fi.ModTime(),
		"mode":     fi.Mode(),
	}
	write(m)
}

func writeText(m map[string]interface{}) {
	for k, v := range m {
		fmt.Printf("%s: %v\n", k, v)
	}
}

func writeJSON(m map[string]interface{}) {
	m["mode"] = m["mode"].(os.FileMode).String()
	json.NewEncoder(os.Stdout).Encode(m)
}

```

Then

```
$ ./finfo finfo.go
mode: -rw-r--r--
size: 783
dir: false
modified: 2019-11-27 11:49:03.280857863 +0200 IST
$ ./finfo -json finfo.go
{"dir":false,"mode":"-rw-r--r--","modified":"2019-11-27T11:49:03.280857863+02:00","size":783}
```

## Progress

Some operations can take long time, one way to make them faster is not by
optimising the code but by showing a spinner/progress bar. Don't believe me,
here's an excerpt from [Nielsen
research](https://www.nngroup.com/articles/progress-indicators/)

> people who saw the moving feedback bar experienced higher satisfaction and
> were willing to wait on average 3 times longer than those who did not see any
> progress indicators.

### Spinner

Adding a spinner does not require any special packages:

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var spinChars = `|/-\`

type Spinner struct {
	message string
	i       int
}

func NewSpinner(message string) *Spinner {
	return &Spinner{message: message}
}

func (s *Spinner) Tick() {
	fmt.Printf("%s %c \r", s.message, spinChars[s.i])
	s.i = (s.i + 1) % len(spinChars)
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func main() {
	flag.Parse()
	s := NewSpinner("working...")
	for i := 0; i < 100; i++ {
		if isTTY() {
			s.Tick()
		}
		time.Sleep(100 * time.Millisecond)
	}

}
```

Run it and you'll see a small spinner going.


### Progress Bar

For a progress bar, you'll probably need an external package such as
`github.com/cheggaaa/pb/v3`

```go
package main

import (
	"flag"
	"time"

	"github.com/cheggaaa/pb/v3"
)

func main() {
	flag.Parse()
	count := 100
	bar := pb.StartNew(count)
	for i := 0; i < count; i++ {
		time.Sleep(100 * time.Millisecond)
		bar.Increment()
	}
	bar.Finish()

}
```

Run it and you'll see a nice progress bar.


# Conclusion

It's almost 2020, and command line applications are here to stay. They are the
key to automation and if written well, provide elegant "lego like" components
to build complex flows.

I hope that this article will prompt you to be a good citizen of the command
line nation.

# About the Author

Hi there, I'm Miki, nice to e-meet you ☺. I've been a long time developer and
have been working with Go for about 10 years now. I write code professionally as
a consultant and contribute a lot to open source. Apart from that I'm a [book
author](https://www.amazon.com/Forging-Python-practices-lessons-developing-ebook/dp/B07C1SH5MP),
an author on [LinkedIn
learning](https://www.linkedin.com/learning/search?keywords=miki+tebeka), one of
the organizers of [GopherCon Israel](https://www.gophercon.org.il/) and [an
instructor](https://www.353.solutions/workshops).  Feel free to [drop me a
line](mailto:miki@353solutions.com) and let me know if you learned something
new or if you'd like to learn more.
