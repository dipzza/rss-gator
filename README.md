# Gator
RSS feed aggregator cli built following guided boot.dev course.

## Prerequisites

To run this program, you will need to have the following installed:

- Go
- PostgreSQL v15 or later

For development you'll need the following tools:

- [Goose](https://github.com/pressly/goose) for database migrations
- [SQLC](https://github.com/kyleconroy/sqlc) to generate type-safe code from SQL

## Installation

Install gator with:

```sh
go install github.com/gaba-bouliva/gator@latest
```

### Configuration

You'll need a PostgreSQL database and some configuration.

1. Create a new database and take note of the connection URL. It should like similar to this: `postgres://postgres:@localhost:5432/db-name`

2. Create a configuration file at `~/.gatorconfig.json`:

```json
{
  "db_url": "postgres://postgres:@localhost:5432/db-name?sslmode=disable",
  "current_user_name": ""
}
```
The `db_url` field is your connection string with an additional parameter specified with ?sslmode=disable

3. Run the database migrations:

```sh
goose -dir sql/schema postgres "postgres://postgres:@localhost:5432/db-name" up
```

## Usage

```sh
gator <command> [argument(s)]
```

### Commands

- `gator login <username>`: Logs in as a user.
- `gator register <username>`: Registers a new user.
- `gator reset`: Cleans up the user database.
- `gator users`: Lists all registered users.
- `gator agg <time_between_reqs>`: Starts getting feeds. Specify the time between request with input like: `5s`, `10m`, `1h`.
- `gator addfeed <name> <url>`: Adds a new feed.
- `gator feeds`: Lists all feeds.
- `gator follow <url>`: Logged user follows a feed.
- `gator following`: Lists feeds followed by the logged user.
- `gator unfollow <url>`: Logged user unfollows a feed.
- `gator browse [n_posts]`: Shows the latest n_posts. By default it shows 2 posts.