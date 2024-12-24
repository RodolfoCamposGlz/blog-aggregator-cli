package login

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/RodolfoCamposGlz/internal/command"
	"github.com/RodolfoCamposGlz/internal/state"
)

func HandlerLogin(s *state.State, cmd command.Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("you need to provide at least one command")
	}
	username :=cmd.Args[0]

	_, err := s.DB.GetUserByName(context.Background(), username)
	if err != nil {
		if err == sql.ErrNoRows {
			// User doesn't exist, continue with user creation
			log.Printf("No existing user found with the name '%s'", username)
			os.Exit(1)
		} else {
			// If another error occurs, log and return
			log.Printf("Error checking if user exists: %v", err)
			return err
		}
	}

	err = s.Config.SetUser(username)
	if err != nil {
		return fmt.Errorf("error setting the username: %w", err)
	}
	fmt.Println("User has been set.")
	return nil
}