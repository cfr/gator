package main

import (
    "fmt"
    "internal/config"
    "github.com/cfr/gator/internal/database"
    "errors"
    "os"
    "time"
    "context"
    "html"
    "database/sql"
    "github.com/google/uuid"
)
import _ "github.com/lib/pq"

type state struct {
    C *config.Config
    Db *database.Queries
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
    feed, err := get[RSSFeed](ctx, feedURL)
    if err != nil {
        return nil, err
    }
    feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
    feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
    for i, item := range feed.Channel.Item {
        feed.Channel.Item[i].Title = html.UnescapeString(item.Title)
        feed.Channel.Item[i].Description = html.UnescapeString(item.Description)
    }
    return feed, nil
}

func handlerAgg(s *state, cmd command) error {
    url := "https://www.wagslane.dev/index.xml"
    feed, err := fetchFeed(context.Background(), url)
    if err != nil {
        return err
    }
    fmt.Println(feed.Channel.Title + " : " + feed.Channel.Description)
    for _, item := range feed.Channel.Item {
        fmt.Println()
        fmt.Println(item.Title)
        fmt.Println(item.Description)
    }
    return nil
}

func handlerUsers(s *state, cmd command) error {
    users, err := s.Db.GetUsers(context.Background())
    if (err != nil) {
        return err
    }
    current := s.C.CurrentUserName
    for _, u := range users {
        line := "* " + u.Name
        if u.Name == current {
            line += " (current)"
        }
        fmt.Println(line)
    }
    return nil
}

func handlerReset(s *state, cmd command) error {
    return s.Db.DeleteUsers(context.Background())
}

func handlerLogin(s *state, cmd command) error {
    if len(cmd.args) == 0 {
        return errors.New("No username provided")
    }
    name := cmd.args[0]
    _, err := s.Db.GetUser(context.Background(), name)
    if (err != nil) {
        return err
    }
    err = s.C.SetUser(name)
    if (err != nil) {
        return err
    }
    fmt.Println("User set: " + name)
    return nil
}

func handlerRegister(s *state, cmd command) error {
    if len(cmd.args) == 0 {
        return errors.New("No username provided")
    }

    currentTime := time.Now()
    params := database.CreateUserParams { uuid.New(), currentTime, currentTime, cmd.args[0] }
    usr, err := s.Db.CreateUser(context.Background(), params)
    if (err != nil) {
        return err
    }
    err = s.C.SetUser(cmd.args[0])
    if (err != nil) {
        return err
    }
    fmt.Println("Created user: " + usr.Name)
    return nil
}

func main() {
    cfg, err := config.Read()
    if err != nil {
        fmt.Println(err)
    }

    st := state{}
    st.C = &cfg

    db, err := sql.Open("postgres", st.C.DbUrl)
    st.Db = database.New(db)

    cmds := commands{make(map[string]func(*state, command) error, 0)}
    cmds.register("login", handlerLogin)
    cmds.register("register", handlerRegister)
    cmds.register("reset", handlerReset)
    cmds.register("users", handlerUsers)
    cmds.register("agg", handlerAgg)

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
