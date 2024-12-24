package main

import (
	"fmt"
	"log"
	"os"

	"database/sql"

	"github.com/RodolfoCamposGlz/internal/command"
	"github.com/RodolfoCamposGlz/internal/config"
	"github.com/RodolfoCamposGlz/internal/database"
	"github.com/RodolfoCamposGlz/internal/login"
	"github.com/RodolfoCamposGlz/internal/state"
	"github.com/RodolfoCamposGlz/middleware"
	_ "github.com/lib/pq"
)

const dbURL = "postgres://rodolfocampos:@localhost:5432/gator?sslmode=disable"


func main (){
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	dbQueries := database.New(db)

	cnf, err:= config.Read()
	if err != nil {
		log.Fatalf("Failed to read config:%v", err)
	}
	stateInstance:= state.State {
		Config: cnf,
		DB: dbQueries,
	}

	cmdsInstance := command.Commands{
		Handler: make(map[string]func(*state.State, command.Command) error),
	}
	
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("you should provide a command")
		os.Exit(1)
	}

	cmdName:= args[0]
	cmdArgs:= args[1:]
	cmdInstance := command.Command {
		Name: cmdName,
		Args: cmdArgs,
	}
	cmdsInstance.Register("login", login.HandlerLogin)
	cmdsInstance.Register("register", cmdsInstance.RegisterUser)
	cmdsInstance.Register("reset", cmdsInstance.DeleteUsers)
	cmdsInstance.Register("users", cmdsInstance.ListUsers)
	cmdsInstance.Register("agg",  middleware.MiddlewareLoggedIn(cmdsInstance.Aggregator))
	cmdsInstance.Register("addfeed", middleware.MiddlewareLoggedIn(cmdsInstance.AddFeed))
	cmdsInstance.Register("feeds", cmdsInstance.GetFeeds)
	cmdsInstance.Register("follow",middleware.MiddlewareLoggedIn(cmdsInstance.Follow))
	cmdsInstance.Register("following", middleware.MiddlewareLoggedIn(cmdsInstance.Following))
	cmdsInstance.Register("unfollow", middleware.MiddlewareLoggedIn(cmdsInstance.UnFollow))
	cmdsInstance.Register("browse", cmdsInstance.Browse)

	err = cmdsInstance.Run(&stateInstance, cmdInstance)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

}