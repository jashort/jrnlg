package cli

import (
	"github.com/jashort/jrnlg/internal"
)

// App coordinates CLI operations
type App struct {
	storage *internal.FileSystemStorage
	config  *internal.Config
}

// NewApp creates a new CLI application
func NewApp(storage *internal.FileSystemStorage, config *internal.Config) *App {
	return &App{
		storage: storage,
		config:  config,
	}
}
