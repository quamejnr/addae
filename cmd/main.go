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
	if err := db.RunMigrations(database); err != nil {
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

// 	// Command line flags
// 	createProjectCmd := flag.NewFlagSet("create-project", flag.ExitOnError)
// 	projectName := createProjectCmd.String("name", "", "Project name")
// 	projectDesc := createProjectCmd.String("desc", "", "Project description")
//
// 	createTaskCmd := flag.NewFlagSet("create-task", flag.ExitOnError)
// 	taskProjectID := createTaskCmd.Int("project", 0, "Project ID")
// 	taskTitle := createTaskCmd.String("title", "", "Task title")
// 	taskDesc := createTaskCmd.String("desc", "", "Task description")
//
// 	createLogCmd := flag.NewFlagSet("create-log", flag.ExitOnError)
// 	logProjectID := createLogCmd.Int("project", 0, "Project ID")
// 	logTitle := createLogCmd.String("title", "", "Log title")
//
// 	if len(os.Args) < 2 {
// 		fmt.Println("Expected subcommands: create-project, create-task, create-log, list-projects")
// 		os.Exit(1)
// 	}
//
// 	switch os.Args[1] {
// 	case "create-project":
// 		createProjectCmd.Parse(os.Args[2:])
// 		if err := createProject(db, *projectName, *projectDesc); err != nil {
// 			log.Fatal(err)
// 		}
// 		fmt.Println("Project created successfully")
//
// 	case "create-task":
// 		createTaskCmd.Parse(os.Args[2:])
// 		if err := createTask(db, *taskProjectID, *taskTitle, *taskDesc); err != nil {
// 			log.Fatal(err)
// 		}
// 		fmt.Println("Task created successfully")
//
// 	case "create-log":
// 		createLogCmd.Parse(os.Args[2:])
// 		if err := createLog(db, *logProjectID, *logTitle); err != nil {
// 			log.Fatal(err)
// 		}
// 		fmt.Println("Log created successfully")
//
// 	case "list-projects":
// 		projects, err := listProjects(db)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		for _, p := range projects {
// 			fmt.Printf("Project %d: %s (%s)\n", p.ID, p.Name, p.Status)
// 			tasks, err := listProjectTasks(db, p.ID)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			for _, t := range tasks {
// 				fmt.Printf("  Task: %s (%s)\n", t.Title, t.Status)
// 			}
// 			logs, err := listProjectLogs(db, p.ID)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			for _, n := range logs {
// 				fmt.Printf("  Log: %s\n", n.Title)
// 			}
// 		}
//
// 	default:
// 		fmt.Println("Expected subcommands: create-project, create-task, create-log, list-projects")
// 		os.Exit(1)
// 	}
