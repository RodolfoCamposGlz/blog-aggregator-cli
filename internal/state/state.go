package state

import (
	"github.com/RodolfoCamposGlz/internal/config"
	"github.com/RodolfoCamposGlz/internal/database"
)

type State struct {
	DB  *database.Queries
	Config *config.Config
}
