package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/beevik/id3"
	"github.com/beevik/prefixtree"
)

var cmds = newCommands([]command{
	{name: "file", description: "Run a file command", commands: newCommands([]command{
		{name: "open", description: "Open a file", handler: onFileOpen},
		{name: "close", description: "Close the open file", handler: onFileClose},
	})},
	{name: "tag", description: "Run a tag command", commands: newCommands([]command{
		{name: "read", description: "Read the next tag from the open file", handler: onTagRead},
		{name: "print", description: "Display the active tag's contents", handler: onTagPrint},
		{name: "dump", description: "Hex dump the active tag", handler: onTagDump},
	})},
	{name: "frame", description: "Find a frame with the given ID", commands: newCommands([]command{
		{name: "activate", description: "Activate a frame with the given ID", handler: onFrameActivate},
		{name: "deactivate", description: "Deactivate the active frame", handler: onFrameDeactivate},
	})},
	{name: "status", description: "Display the current status", handler: onStatus},
	{name: "exit", description: "", handler: onQuit},
	{name: "quit", description: "Exit the application", handler: onQuit},
})

type state struct {
	activeFile          *os.File
	activeFileReader    *bufio.Reader
	activeFileBytesRead int
	activeFilename      string
	activeTag           *id3.Tag
	activeFrame         id3.Frame
}

func (s *state) reset() {
	if s.activeFile != nil {
		s.activeFile.Close()
	}
	*s = state{}
}

func main() {
	args := os.Args
	switch {
	case len(args) == 1:
		startRepl()
	default:
		exec(args[1])
	}
}

func exec(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Command file '%s' not found.\n", filename)
		os.Exit(0)
	}

	c := newConn(file, os.Stdout)
	return repl(c)
}

func startRepl() error {
	c := newConn(os.Stdin, os.Stdout)
	c.interactive = true
	return repl(c)
}

func repl(c *conn) error {
	s := &state{}

	for {
		if c.interactive {
			c.Printf("id3> ")
			c.Flush()
		}

		line, err := c.GetLine()
		if err != nil {
			break
		}

		if !c.interactive {
			c.Printf("id3> %s\n", line)
		}

		r, err := cmds.find(line)
		switch {
		case err == prefixtree.ErrPrefixNotFound:
			c.Println("command not found.")
			continue
		case err == prefixtree.ErrPrefixAmbiguous:
			c.Println("command ambiguous.")
			continue
		case err != nil:
			c.Printf("%v.\n", err)
			continue
		case r.helpText != "":
			c.Printf("%s", r.helpText)
			continue
		case r.cmd == nil:
			continue
		}

		err = r.cmd.handler(c, s, r.args)
		if err != nil {
			break
		}
		c.Printf("\n")
	}

	return nil
}

func onFileOpen(c *conn, s *state, args string) error {
	segments := strings.Split(args, " ")

	if len(segments) < 1 || segments[0] == "" {
		c.Println("invalid filename")
		return nil
	}

	file, err := os.Open(segments[0])
	if err != nil {
		c.Printf("%v\n", err)
		return nil
	}

	s.reset()

	c.Printf("File '%v' opened successfully.\n", segments[0])
	s.activeFile = file
	s.activeFilename = segments[0]
	s.activeFileReader = bufio.NewReader(s.activeFile)
	s.activeFileBytesRead = 0
	return nil
}

func onFileClose(c *conn, s *state, args string) error {
	s.reset()
	c.Println("OK.")
	return nil
}

func onTagRead(c *conn, s *state, args string) error {
	if s.activeFileReader == nil {
		c.Println("ERROR: No active file opened.")
		return nil
	}

	p, err := s.activeFileReader.Peek(10)
	if err != nil {
		c.Printf("ERROR: %v\n", err)
		return nil
	}

	_, _, err = id3.PeekTag(p)
	if err != nil {
		c.Printf("ERROR: %v\n", err)
		return nil
	}

	t := &id3.Tag{}
	n, err := t.ReadFrom(s.activeFileReader)
	s.activeFileBytesRead += int(n)
	if err == id3.ErrInvalidTag {
		c.Println("ERROR: No valid tag discovered.")
		return nil
	}
	if err != nil {
		c.Printf("ERROR: %v\n", err)
		return nil
	}

	c.Printf("Version 2.%d tag successfully read (%d bytes).\n", t.Version, n)
	s.activeTag = t

	return nil
}

func onTagPrint(c *conn, s *state, args string) error {
	if s.activeTag == nil {
		c.Println("ERROR: No active tag.")
		return nil
	}

	outputTag(c, s.activeTag)

	for _, f := range s.activeTag.Frames {
		outputFrame(c, f)
	}

	return nil
}

func onTagDump(c *conn, s *state, args string) error {
	if s.activeTag == nil {
		c.Println("ERROR: No active tag.")
		return nil
	}

	buf := bytes.NewBuffer([]byte{})
	_, err := s.activeTag.WriteTo(buf)
	if err != nil {
		c.Printf("ERROR: %v\n", err)
		return nil
	}

	hexdump(c, buf.Bytes())
	return nil
}

func onFrameActivate(c *conn, s *state, args string) error {
	if s.activeTag == nil {
		c.Println("ERROR: No active tag.")
		return nil
	}

	arg := strings.Split(args, " ")

	var found id3.Frame
	for _, f := range s.activeTag.Frames {
		h := id3.HeaderOf(f)
		if h.FrameID == arg[0] {
			found = f
			break
		}
	}

	if found == nil {
		c.Printf("ERROR: Frame '%s' not found.\n", arg[0])
		return nil
	}

	s.activeFrame = found
	c.Printf("Frame '%s' active.\n", arg[0])
	outputFrame(c, s.activeFrame)
	return nil
}

func onFrameDeactivate(c *conn, s *state, args string) error {
	if s.activeFrame == nil {
		c.Println("ERROR: No active tag.")
		return nil
	}

	h := id3.HeaderOf(s.activeFrame)
	id := h.FrameID

	s.activeFrame = nil
	c.Printf("Frame '%s' deactivated.", id)
	return nil
}

func onStatus(c *conn, s *state, args string) error {
	if s.activeFileReader == nil {
		c.Println("No active file.")
	} else {
		c.Println("Active file:")
		c.Printf("  Name:       %s\n", s.activeFilename)
		c.Printf("  Bytes read: %d\n", s.activeFileBytesRead)
		c.Printf("\n")
	}

	if s.activeTag == nil {
		c.Println("No active tag.")
	} else {
		c.Println("Active tag:")
		c.Printf("  version: 2.%d\n", s.activeTag.Version)
		c.Printf("  size:    %d bytes\n", s.activeTag.Size+10)
		c.Printf("  padding: %d bytes\n", s.activeTag.Padding)
		if (s.activeTag.Flags & id3.TagFlagHasCRC) != 0 {
			c.Printf("  crc:     0x%08x\n", s.activeTag.CRC)
		}
		c.Printf("  frames:  %d\n", len(s.activeTag.Frames))
		for _, f := range s.activeTag.Frames {
			h := id3.HeaderOf(f)
			c.Printf("    %s: %d bytes\n", h.FrameID, h.Size)
		}
		c.Printf("\n")
	}

	if s.activeFrame == nil {
		c.Println("No active frame.")
	} else {
		c.Println("Active frame:")
		outputFrame(c, s.activeFrame)
	}

	return nil
}

func onQuit(c *conn, s *state, args string) error {
	return errors.New("quitting")
}

func stripLeadingWhitespace(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return s[i:]
		}
	}
	return ""
}

func outputTag(c *conn, tag *id3.Tag) {
	c.Printf("Version: 2.%d\n", tag.Version)
	c.Printf("Size: %d bytes\n", tag.Size+10)
	if (tag.Flags & id3.TagFlagHasCRC) != 0 {
		c.Printf("CRC: 0x%08x\n", tag.CRC)
	}
	if tag.Padding > 0 {
		c.Printf("Pad: %d bytes\n", tag.Padding)
	}
}

func outputFrame(c *conn, ff id3.Frame) {
	hdr := id3.HeaderOf(ff)
	c.Printf("  [size=0x%04x] %s", hdr.Size+10, hdr.FrameID)

	switch f := ff.(type) {
	case *id3.FrameUnknown:
		c.Printf(": (%d bytes)", len(f.Data))
	case *id3.FrameAttachedPicture:
		c.Printf(": #%d %s[%s] (%d bytes)", f.PictureType, f.Description, f.MimeType, len(f.Data))
	case *id3.FrameText:
		c.Printf(": %s", strings.Join(f.Text, " - "))
	case *id3.FrameTextCustom:
		c.Printf(": %s -> %s", f.Description, f.Text)
	case *id3.FrameComment:
		c.Printf(": %s -> %s", f.Description, f.Text)
	case *id3.FrameURL:
		c.Printf(": %s", f.URL)
	case *id3.FrameURLCustom:
		c.Printf(": %s -> %s", f.Description, f.URL)
	case *id3.FrameUniqueFileID:
		c.Printf(": %s -> %s", f.Owner, f.Identifier)
	case *id3.FrameLyricsUnsync:
		c.Printf(": [%s:%s] %s", f.Language, f.Descriptor, f.Text)
	case *id3.FrameLyricsSync:
		c.Printf(": [%s:%s] %d syncs", f.Language, f.Descriptor, len(f.Sync))
		for _, s := range f.Sync {
			c.Printf("\n    %d: %s", s.TimeStamp, strings.Replace(s.Text, "\n", "<CR>", -1))
		}
	case *id3.FramePrivate:
		data := f.Data
		if len(data) > 32 {
			data = data[:32]
		}
		c.Printf(": %s %v (%d bytes)", f.Owner, data, len(f.Data))
	case *id3.FramePlayCount:
		c.Printf(": %d", f.Counter)
	case *id3.FramePopularimeter:
		c.Printf(": %s (%d) %d", f.Email, f.Rating, f.Counter)
	}
	c.Printf("\n")
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
	lyr.AddSync(3000, "This ")
	lyr.AddSync(1000, "is ")
	lyr.AddSync(2001, "a song.")
	tag.Frames = append(tag.Frames, lyr)

	playcount := id3.NewFramePlayCount(0x1234567890aabbcc)
	tag.Frames = append(tag.Frames, playcount)

	title := id3.NewFrameText(id3.FrameTypeTextSongTitle, "Yellow Submarine")
	tag.Frames = append(tag.Frames, title)

	tx := id3.NewFrameTextCustom("label", "content")
	tag.Frames = append(tag.Frames, tx)

	priv := id3.NewFramePrivate("owner", []byte{0, 1, 2, 3})
	tag.Frames = append(tag.Frames, priv)

	sp := id3.NewFrameAudioSeekPointIndex(0, 1000)
	sp.AddIndexOffset(100)
	sp.AddIndexOffset(2)
	sp.AddIndexOffset(951)
	sp.AddIndexOffset(800)
	sp.AddIndexOffset(401)
	tag.Frames = append(tag.Frames, sp)

	buf := bytes.NewBuffer([]byte{})
	tag.WriteTo(buf)

	c := newConn(os.Stdin, os.Stdout)
	hexdump(c, buf.Bytes())
}

func hexdump(c *conn, b []byte) {
	c.Printf("var b = []byte{\n")

	for i := 0; i < len(b); i += 8 {
		r := i + 8
		if r > len(b) {
			r = len(b)
		}

		c.Printf("\t")

		var j int
		for j = i; j < r-1; j++ {
			c.Printf("0x%02x, ", b[j])
		}
		c.Printf("0x%02x,\n", b[j])
	}

	c.Printf("}\n")
}
