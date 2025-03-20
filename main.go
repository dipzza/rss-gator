package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/dipzza/rss-gator/internal/config"
	"github.com/dipzza/rss-gator/internal/database"

	_ "github.com/lib/pq"
)

type state struct {
	db *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("%s is not a gator command.", cmd.name)
	}

	return f(s, cmd)
}

func main () {
	c := commands{
		handlers: map[string]func(*state, command) error{
			"login": loginHandler,
			"register": registerHandler,
			"reset": resetHandler,
			"users": usersHandler,
			"agg": aggHandler,
			"addfeed": middlewareLoggedIn(addfeedHandler),
			"feeds": feedsHandler,
			"follow": middlewareLoggedIn(followHandler),
			"following": middlewareLoggedIn(followingHandler),
			"unfollow": middlewareLoggedIn(unfollowHandler),
			"browse": middlewareLoggedIn(browseHandler),
		},
	}

	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	s := state{
		cfg: &cfg,
		db: dbQueries,
	}

	if len(os.Args) < 2 {
		fmt.Println("Error: No command selected.")
		os.Exit(1)
	}
	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}
	if err := c.run(&s, cmd); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

