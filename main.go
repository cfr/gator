package main

import (
    "fmt"
    "internal/config"
    "internal/database"
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

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
    return func(s *state, cmd command) error {
        user, err := s.Db.GetUser(context.Background(), s.C.CurrentUserName)
        if err != nil {
            return err
        }
        return handler(s, cmd, user)
    }
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
    if err != nil {
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
    err := s.Db.DeleteUsers(context.Background())
    if err != nil {
        return err
    }
    err = s.Db.DeleteFeeds(context.Background())
    if err != nil {
        return err
    }
    return s.Db.DeleteFeedFollows(context.Background())
}

func handlerLogin(s *state, cmd command) error {
    if len(cmd.args) < 1 {
        return errors.New("No username provided")
    }
    name := cmd.args[0]
    _, err := s.Db.GetUser(context.Background(), name)
    if err != nil {
        return err
    }
    err = s.C.SetUser(name)
    if err != nil {
        return err
    }
    fmt.Println("User set: " + name)
    return nil
}

func handlerRegister(s *state, cmd command) error {
    if len(cmd.args) < 1 {
        return errors.New("No username provided")
    }

    currentTime := time.Now()
    params := database.CreateUserParams { uuid.New(), currentTime, currentTime, cmd.args[0] }
    usr, err := s.Db.CreateUser(context.Background(), params)
    if err != nil {
        return err
    }
    err = s.C.SetUser(cmd.args[0])
    if err != nil {
        return err
    }
    fmt.Println("Created user: " + usr.Name)
    return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
    if len(cmd.args) < 2 {
        return errors.New("No feed name and url provided")
    }

    currentTime := time.Now()
    params := database.CreateFeedParams { uuid.New(), currentTime, currentTime, cmd.args[0], cmd.args[1], user.ID }
    feed, err := s.Db.CreateFeed(context.Background(), params)
    if err != nil {
        return err
    }
    followParams := database.CreateFeedFollowParams { uuid.New(), currentTime, currentTime, user.ID, feed.ID }
    ff, err := s.Db.CreateFeedFollow(context.Background(), followParams)
    if err != nil {
        return err
    }
    fmt.Println("Created and followed feed: " + ff.FeedName + ", " + feed.Url)
    return nil
}

func handlerFeeds(s *state, cmd command) error {
    feeds, err := s.Db.GetFeeds(context.Background())
    if err != nil {
        return err
    }
    for _, f := range feeds {
        user, err := s.Db.GetUserByID(context.Background(), f.UserID)
        if err != nil {
            return err
        }
        line := "* " + user.Name + " : " + f.Name + " : " + f.Url
        fmt.Println(line)
    }
    return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
    if len(cmd.args) < 1 {
        return errors.New("No feed url provided")
    }
    feed, err := s.Db.GetFeedByURL(context.Background(), cmd.args[0])
    if err != nil {
        return err
    }

    currentTime := time.Now()

    params := database.CreateFeedFollowParams { uuid.New(), currentTime, currentTime, user.ID, feed.ID }
    follow, err := s.Db.CreateFeedFollow(context.Background(), params)
    if err != nil {
        return err
    }
    line := "Followed " + follow.FeedName + " for " + follow.UserName + " : " + feed.Url
    fmt.Println(line)
    return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
    if len(cmd.args) < 1 {
        return errors.New("No feed url provided")
    }
    feed, err := s.Db.GetFeedByURL(context.Background(), cmd.args[0])
    if err != nil {
        return err
    }

    params := database.DeleteFeedFollowParams { user.ID, feed.ID }
    err = s.Db.DeleteFeedFollow(context.Background(), params)
    if err != nil {
        return err
    }
    line := "Unfollowed " + feed.Url
    fmt.Println(line)
    return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
    feed_follows, err := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
    if err != nil {
        return err
    }

    for _, f := range feed_follows {
        fmt.Println("* " + f.FeedName)
    }
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
    cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
    cmds.register("feeds", handlerFeeds)
    cmds.register("follow", middlewareLoggedIn(handlerFollow))
    cmds.register("following", middlewareLoggedIn(handlerFollowing))
    cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))

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
