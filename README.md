# gator

A CLI RSS feed aggregator written in Go.

## Prerequisites

- [Go](https://go.dev/dl/) 1.26+
- [PostgreSQL](https://www.postgresql.org/download/)

## Installation

```bash
go install github.com/cfr/gator@latest
```

## Setup

Create a config file at `~/.gatorconfig.json` with your Postgres connection string:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Then apply the database migrations:

```bash
goose -dir sql/schema postgres "postgres://username:password@localhost:5432/gator?sslmode=disable" up
```

## Usage

Register a user (this also logs you in):

```bash
gator register alice
```

Add a feed (must be logged in):

```bash
gator addfeed "Hacker News" https://news.ycombinator.com/rss
```

Start the aggregator in one terminal (it fetches posts on an interval):

```bash
gator agg 1m
```

Browse the latest posts from feeds you follow in another terminal:

```bash
gator browse
```

Other commands:

- `gator login <username>` — switch users
- `gator users` — list all users
- `gator feeds` — list all feeds
- `gator follow <url>` / `gator unfollow <url>` — manage feed follows
- `gator following` — list feeds you follow
- `gator reset` — wipe all data

## Development notes

After editing files in `sql/queries/` or `sql/schema/`, regenerate the Go database code:

```bash
sqlc generate
```

To re-run all migrations from scratch (drops everything, then re-applies):

```bash
goose -dir sql/schema postgres "$DB_URL" down
goose -dir sql/schema postgres "$DB_URL" up
```

