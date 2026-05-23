package main

import (
    "fmt"
    "internal/config"
    "errors"
    "os"
)

type state struct {
    C *config.Config
}

func handlerLogin(s *state, cmd command) error {
    if len(cmd.args) == 0 {
        return errors.New("No username provided")
    }
    err := s.C.SetUser(cmd.args[0])
    if (err != nil) {
        return err
    }
    fmt.Println("User set: " + cmd.args[0])
    return nil
}

func main() {
    cfg, err := config.Read()
    if err != nil {
        fmt.Println(err)
    }

    st := state{}
    st.C = &cfg

    cmds := commands{make(map[string]func(*state, command) error, 0)}
    cmds.register("login", handlerLogin)

    if len(os.Args) < 2 {
        fmt.Println("Missing command")
        os.Exit(1)
    }

    name := os.Args[1]
    args := os.Args[2:]
    cmd := command{name, args}

    err = cmds.run(&st, cmd)

    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
