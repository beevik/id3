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

	for {
		var tag id3.Tag
		_, err = tag.ReadFrom(file)
		if err == id3.ErrInvalidTag {
			break
		}
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Version: 2.%d\n", tag.Version)
		fmt.Printf("Size: %d bytes\n", tag.Size)
		if (tag.Flags & id3.TagFlagHasCRC) != 0 {
			fmt.Printf("CRC: 0x%08x\n", tag.CRC)
		}
		if tag.Padding > 0 {
			fmt.Printf("Pad: %d bytes\n", tag.Padding)
		}

		for _, f := range tag.Frames {
			fmt.Printf("[size=0x%04x] %s", f.Header.Size+10, f.Header.ID)
			switch ff := f.Payload.(type) {
			case *id3.FramePayloadUnknown:
				fmt.Printf(": (%d bytes)", len(ff.Data))
			case *id3.FramePayloadAPIC:
				fmt.Printf(": #%d %s[%s] (%d bytes)", ff.Type, ff.Description, ff.MimeType, len(ff.Data))
			case *id3.FramePayloadText:
				fmt.Printf(": %s", strings.Join(ff.Text, " - "))
			case *id3.FramePayloadTXXX:
				fmt.Printf(": %s -> %s", ff.Description, ff.Text)
			case *id3.FramePayloadCOMM:
				fmt.Printf(": %s -> %s", ff.Description, ff.Text)
			case *id3.FramePayloadURL:
				fmt.Printf(": %s", ff.URL)
			case *id3.FramePayloadWXXX:
				fmt.Printf(": %s -> %s", ff.Description, ff.URL)
			case *id3.FramePayloadUFID:
				fmt.Printf(": %s -> %s", ff.Owner, ff.Identifier)
			case *id3.FramePayloadUSLT:
				fmt.Printf(": [%s:%s] %s", ff.Language, ff.Descriptor, ff.Text)
			case *id3.FramePayloadSYLT:
				fmt.Printf(": [%s:%s] %d syncs", ff.Language, ff.Descriptor, len(ff.Sync))
				for _, s := range ff.Sync {
					fmt.Printf("\n  %d: %s", s.TimeStamp, s.Text)
				}
			case *id3.FramePayloadPRIV:
				fmt.Printf(": %s %v (%d bytes)", ff.Owner, ff.Data, len(ff.Data))
			case *id3.FramePayloadPCNT:
				fmt.Printf(": %d", ff.Count)
			case *id3.FramePayloadPOPM:
				fmt.Printf(": %s (%d) %d", ff.Email, ff.Rating, ff.Count)
			}
			fmt.Printf("\n")
		}
	}
}

func usage() {
	fmt.Println(`Syntax: id3repl [file]`)
	os.Exit(0)
}
