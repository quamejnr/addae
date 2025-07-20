
package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quamejnr/addae/internal/service"
)

func TestUpdateListView(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	model, err := NewModel(mockService)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test navigating to the create view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	newModel, _ := model.Update(msg)
	model = newModel.(*Model)

	if model.GetState() != createView {
		t.Errorf("expected state to be createView, got %v", model.GetState())
	}
	if model.form == nil {
		t.Error("expected form to be initialized, but it was nil")
	}
}

func TestUpdateProjectView(t *testing.T) {
	mockService := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	model, err := NewModel(mockService)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Select a project to enter project view
	model.CoreModel.SelectProject(0)

	// Test navigating to the update view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")}
	newModel, _ := model.Update(msg)
	model = newModel.(*Model)

	if model.GetState() != updateView {
		t.Errorf("expected state to be updateView, got %v", model.GetState())
	}
	if model.form == nil {
		t.Error("expected form to be initialized, but it was nil")
	}

	// Test navigating to the create task view
	model.CoreModel.GoToProjectView() // Go back to project view first
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	newModel, _ = model.Update(msg)
	model = newModel.(*Model)

	if model.GetState() != createTaskView {
		t.Errorf("expected state to be createTaskView, got %v", model.GetState())
	}
	if model.form == nil {
		t.Error("expected form to be initialized, but it was nil")
	}
}

func TestHandleFormAbort(t *testing.T) {
	// Test aborting from create view
	mockServiceCreate := &MockService{}
	model, _ := NewModel(mockServiceCreate)
	model.CoreModel.GoToCreateView()
	model.handleFormAbort("create")
	if model.GetState() != listView {
		t.Errorf("expected state to be listView after aborting create, got %v", model.GetState())
	}

	// Test aborting from update view
	mockServiceUpdate := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	model, _ = NewModel(mockServiceUpdate)
	model.CoreModel.SelectProject(0)
	model.CoreModel.GoToUpdateView()
	model.handleFormAbort("update")
	if model.GetState() != projectView {
		t.Errorf("expected state to be projectView after aborting update, got %v", model.GetState())
	}

	// Test aborting from createTask view
	mockServiceCreateTask := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	model, _ = NewModel(mockServiceCreateTask)
	model.CoreModel.SelectProject(0)
	model.CoreModel.GoToCreateTaskView()
	model.handleFormAbort("createTask")
	if model.GetState() != projectView {
		t.Errorf("expected state to be projectView after aborting createTask, got %v", model.GetState())
	}

	// Test aborting from createLog view
	mockServiceCreateLog := &MockService{
		projects: []service.Project{{ID: 1, Name: "Test Project"}},
	}
	model, _ = NewModel(mockServiceCreateLog)
	model.CoreModel.SelectProject(0)
	model.CoreModel.GoToCreateLogView()
	model.handleFormAbort("createLog")
	if model.GetState() != projectView {
		t.Errorf("expected state to be projectView after aborting createLog, got %v", model.GetState())
	}
}
