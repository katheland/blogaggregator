package main

import _ "github.com/lib/pq"

import (
	"fmt"
	"internal/config"
	"internal/database"
	"errors"
	"os"
	"database/sql"
)

// holds the config
type state struct {
	config* config.Config
	database* database.Queries
}

// a command's name and arguments
type command struct {
	name string
	arguments []string
}

// a map of commands
type commands struct {
	handlers map[string]func(*state, command) error
}

// register a command
func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

// run a command
func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return errors.New("command does not exist")
	}
	return handler(s, cmd)
}

func main() {
	// read the config, open the database, and save them in a state variable
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	s := state{config: &cfg, database: dbQueries,}

	// register the commands
	c := commands{handlers: make(map[string]func(*state, command) error),}
	registerHandlers(c)

	// call the command that was called
	if len(os.Args) < 2 {
		fmt.Println("not enough arguments provided")
		os.Exit(1)
	}
	name := os.Args[1]
	args := os.Args[2:]
	cmd := command{name: name, arguments: args,}
	err = c.run(&s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}