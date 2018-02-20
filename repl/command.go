package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/beevik/prefixtree"
)

type handlerFunc func(c *conn, s *state, args string) error

type command struct {
	name        string
	description string
	handler     handlerFunc
	commands    *commands
}

type commands struct {
	list    []command
	tree    *prefixtree.Tree
	context string
}

type commandResult struct {
	cmd      *command
	args     string
	helpText string
}

func newCommands(list []command) *commands {
	c := &commands{
		list: list,
		tree: prefixtree.New(),
	}
	for i, cc := range c.list {
		c.tree.Add(cc.name, &c.list[i])
	}
	return c
}

func (c *commands) find(line string) (commandResult, error) {
	ss := strings.SplitN(line, " ", 2)

	var cmd, args string
	switch len(ss) {
	case 1:
		cmd = ss[0]
	case 2:
		cmd, args = ss[0], stripLeadingWhitespace(ss[1])
	}

	if cmd == "help" || cmd == "?" {
		lines := []string{"Commands:\n"}
		for _, c := range c.list {
			if c.description != "" {
				line := fmt.Sprintf("  %-15s  %s\n", c.name, c.description)
				lines = append(lines, line)
			}
		}
		return commandResult{helpText: strings.Join(lines, "")}, nil
	}

	ci, err := c.tree.Find(cmd)
	if err != nil {
		return commandResult{}, err
	}

	cc := ci.(*command)
	switch {
	case cc.handler != nil:
		return commandResult{cmd: cc, args: args}, nil
	case cc.commands != nil:
		if args == "" {
			return commandResult{}, errors.New("command missing argument")
		}
		return cc.commands.find(args)
	}

	return commandResult{}, errors.New("command not found")
}
