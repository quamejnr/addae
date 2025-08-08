package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"

	"github.com/quamejnr/addae/internal/service"
	"github.com/quamejnr/addae/internal/ui"

	"github.com/quamejnr/addae/internal/db"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

//go:embed internal/db/migrations/*.sql
var migrationsFS embed.FS

func main() {
	var showVersion bool

	flag.BoolVar(&showVersion, "version", false, "Print version information")
	flag.Parse()

	if showVersion {
		fmt.Printf("addae version %s\n", version)
		os.Exit(0)
	}

	// Initialize database
	database, err := db.InitDB("")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer database.Close()

	// Get the migrations subdirectory
	migrations, err := fs.Sub(migrationsFS, "internal/db/migrations")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Run migrations with embedded files
	if err := db.RunMigrationsFromFS(database, migrations); err != nil {
		fmt.Println(err)
		return
	}

	// Initialize service with database
	svc := service.NewService(database)

	// Initialize TUI
	model, err := ui.NewModel(svc)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Start TUI program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		return
	}
}
