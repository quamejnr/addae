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