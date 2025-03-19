package main

import (
	"internal/database"
	"database/sql"
	"context"
)

// middleware: handle login checks
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func (st *state, co command) error {
		currentUser, err := st.database.GetUser(context.Background(), sql.NullString{String: st.config.CurrentUserName, Valid: true,})
		if err != nil {
			return err
		}
		return handler(st, co, currentUser)
	}
}