package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/beevik/id3"
)

func main() {

	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		args = make([]string, 1)
		args[0] = "file.mp3"
		//usage()
	}

	file, err := os.Open(args[0])
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	var tag id3.Tag
	_, err = tag.ReadFrom(file)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	for _, f := range tag.Frames {
		fmt.Printf("Frame %s", f.Header.ID)
		switch ff := f.Payload.(type) {
		case *id3.FramePayloadAPIC:
			fmt.Printf(": #%d %s[%s] (%d bytes)", ff.Type, ff.Description, ff.MimeType, len(ff.Data))
		case *id3.FramePayloadText:
			fmt.Printf(": %s", strings.Join(ff.Text, " - "))
		case *id3.FramePayloadTXXX:
			fmt.Printf(": %s -> %s", ff.Description, ff.Text)
		case *id3.FramePayloadUFID:
			fmt.Printf(": %s -> %s", ff.Owner, ff.Identifier)
		case *id3.FramePayloadUSLT:
			fmt.Printf(": [%s:%s] %s", ff.Language, ff.Descriptor, ff.Text)
		case *id3.FramePayloadUnknown:
			fmt.Printf(": (%d bytes)", len(ff.Data))
		}
		fmt.Printf("\n")
	}
}

func usage() {
	fmt.Println(`Syntax: id3repl [file]`)
	os.Exit(0)
}
