package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/RodolfoCamposGlz/internal/command"
	"github.com/RodolfoCamposGlz/internal/database"
	"github.com/RodolfoCamposGlz/internal/state"
)

// middlewareLoggedIn wraps a handler to ensure that a user is logged in.
func MiddlewareLoggedIn(handler func(s *state.State, cmd command.Command, user database.User) error) func(*state.State, command.Command) error {
	return func(s *state.State, cmd command.Command) error {
		// Retrieve the current user from the state (assuming the username is stored in state)
		currentUserName := s.Config.CurrentUserName

		// Check if a current user is set
		if currentUserName == "" {
			return fmt.Errorf("no user is logged in")
		}

		// Fetch the user from the database
		user, err := s.DB.GetUserByName(context.Background(), currentUserName)
		if err != nil {
			// Handle case where the user is not found or other database error
			if err == sql.ErrNoRows {
				log.Printf("No user found with the name '%s'.", currentUserName)
			} else {
				log.Printf("Error retrieving user: %v", err)
			}
			return fmt.Errorf("failed to retrieve user: %v", err)
		}

		// Call the original handler, passing the user as an argument
		return handler(s, cmd, user)
	}
}
