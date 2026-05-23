package main

import (
    "errors"
)

type command struct {
    name string
    args []string
}

type commands struct {
    cs map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
    handler, ok := c.cs[cmd.name]
    if !ok {
        return errors.New("Unknown command " + cmd.name)
    }
    return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
    c.cs[name] = f
}
