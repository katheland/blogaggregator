package main

import _ "github.com/lib/pq"

import (
	"fmt"
	"internal/config"
	"internal/database"
	"errors"
	"os"
	"database/sql"
	"context"
	"github.com/google/uuid"
	"time"
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

// command: log in
func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return errors.New("login requires a username")
	}
	_, err := s.database.GetUser(context.Background(), sql.NullString{String: cmd.arguments[0], Valid: true,})
	if err != nil {
		return errors.New("username does not exist")
	}
	err = s.config.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("user successfully set")
	return nil
}

// command: register a user
func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		return errors.New("register requires a username")
	}
	user, err := s.database.CreateUser(
		context.Background(), 
		database.CreateUserParams{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name: sql.NullString{String: cmd.arguments[0], Valid: true,},
		})
	if err != nil {
		return errors.New("username already exists")
	}
	err = s.config.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("user successfully registered and set")
	fmt.Println(fmt.Sprintf("%v %v %v %v", user.ID, user.CreatedAt, user.UpdatedAt, user.Name))
	return nil
}

// command: reset the database (maybe deregister this one later for "production")
func handlerReset(s *state, cmd command) error {
	err := s.database.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting the users: %v", err)
	}
	fmt.Println("users deleted successfully")
	return nil
}

// command: list all users
func handlerUsers(s *state, cmd command) error {
	users, err := s.database.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all users: %v", err)
	}
	current_user := s.config.CurrentUserName
	for _, u := range users {
		n := u.String
		if n == current_user {
			fmt.Println("* " + n + " (current)")
		} else {
			fmt.Println("* " + n)
		}
	}
	return nil
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
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)

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