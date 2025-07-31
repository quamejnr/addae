package main

import (
	"fmt"

	"github.com/quamejnr/addae/internal/service"
	"github.com/quamejnr/addae/internal/ui"

	"github.com/quamejnr/addae/internal/db"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize database
	database, err := db.InitDB("./addae.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer database.Close()

	// run migrations
	if err := db.RunMigrations(database, "./internal/db/migrations/"); err != nil {
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

