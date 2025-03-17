package main

import (
	"fmt"
	"internal/config"
	"errors"
	"os"
)

// holds the config
type state struct {
	config* config.Config
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

// login command
func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return errors.New("login requires a username")
	}
	err := s.config.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("user successfully set")
	return nil
}

func main() {
	// read the config and save it in a state variable
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	s := state{config: &cfg,}

	// register the commands
	c := commands{handlers: make(map[string]func(*state, command) error),}
	c.register("login", handlerLogin)

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