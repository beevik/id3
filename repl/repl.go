package main

import (
	"bytes"
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

	writeTag()

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
		fmt.Printf("Size: %d bytes\n", tag.Size+10)
		if (tag.Flags & id3.TagFlagHasCRC) != 0 {
			fmt.Printf("CRC: 0x%08x\n", tag.CRC)
		}
		if tag.Padding > 0 {
			fmt.Printf("Pad: %d bytes\n", tag.Padding)
		}

		for _, h := range tag.Frames {
			hdr := id3.HeaderOf(h)
			fmt.Printf("[size=0x%04x] %s", hdr.Size+10, hdr.FrameID)

			switch f := h.(type) {
			case *id3.FrameUnknown:
				fmt.Printf(": (%d bytes)", len(f.Data))
			case *id3.FrameAttachedPicture:
				fmt.Printf(": #%d %s[%s] (%d bytes)", f.PictureType, f.Description, f.MimeType, len(f.Data))
			case *id3.FrameText:
				fmt.Printf(": %s", strings.Join(f.Text, " - "))
			case *id3.FrameTextCustom:
				fmt.Printf(": %s -> %s", f.Description, f.Text)
			case *id3.FrameComment:
				fmt.Printf(": %s -> %s", f.Description, f.Text)
			case *id3.FrameURL:
				fmt.Printf(": %s", f.URL)
			case *id3.FrameURLCustom:
				fmt.Printf(": %s -> %s", f.Description, f.URL)
			case *id3.FrameUniqueFileID:
				fmt.Printf(": %s -> %s", f.Owner, f.Identifier)
			case *id3.FrameLyricsUnsync:
				fmt.Printf(": [%s:%s] %s", f.Language, f.Descriptor, f.Text)
			case *id3.FrameLyricsSync:
				fmt.Printf(": [%s:%s] %d syncs", f.Language, f.Descriptor, len(f.Sync))
				for _, s := range f.Sync {
					fmt.Printf("\n  %d: %s", s.TimeStamp, s.Text)
				}
			case *id3.FramePrivate:
				fmt.Printf(": %s %v (%d bytes)", f.Owner, f.Data, len(f.Data))
			case *id3.FramePlayCount:
				fmt.Printf(": %d", f.Counter)
			case *id3.FramePopularimeter:
				fmt.Printf(": %s (%d) %d", f.Email, f.Rating, f.Counter)
			}
			fmt.Printf("\n")
		}
	}
}

func usage() {
	fmt.Println(`Syntax: id3repl [file]`)
	os.Exit(0)
}

func writeTag() {
	tag := id3.Tag{Version: id3.Version2_4}
	tag.Flags |= id3.TagFlagHasCRC

	com := id3.NewFrameComment("eng", "foo", "comment")
	com.Header.SetGroupID(0x90)
	com.Header.SetEncryptMethod(0xf0)
	com.Header.SetFlag(id3.FrameFlagCompressed, true)
	com.Header.SetFlag(id3.FrameFlagHasDataLength, true)
	tag.Frames = append(tag.Frames, com)

	lyr := id3.NewFrameLyricsSync("eng", "lyrics", id3.TimeStampMilliseconds, id3.LyricContentTypeTranscription)
	lyr.AddSync(id3.LyricsSync{Text: "c", TimeStamp: 3})
	lyr.AddSync(id3.LyricsSync{Text: "a", TimeStamp: 1})
	lyr.AddSync(id3.LyricsSync{Text: "b", TimeStamp: 2})
	tag.Frames = append(tag.Frames, lyr)

	playcount := id3.NewFramePlayCount(0x1234567890aabbcc)
	tag.Frames = append(tag.Frames, playcount)

	title := id3.NewFrameText(id3.FrameTypeTextSongTitle, "Yellow Submarine")
	tag.Frames = append(tag.Frames, title)

	priv := id3.NewFramePrivate("owner", []byte{0, 1, 2, 3})
	tag.Frames = append(tag.Frames, priv)

	sp := id3.NewFrameAudioSeekPointIndex(0, 1000)
	sp.AddIndexOffset(100)
	sp.AddIndexOffset(2)
	sp.AddIndexOffset(951)
	sp.AddIndexOffset(800)
	sp.AddIndexOffset(400)

	tag.Frames = append(tag.Frames, sp)

	buf := bytes.NewBuffer([]byte{})
	tag.WriteTo(buf)

	//hexdump.Dump(buf.Bytes(), hexdump.FormatGo, os.Stdout)
}
