package ui

import (
	"testing"
	"time"

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

func TestGetVisualTask(t *testing.T) {
	now := time.Now()
	tasks := []service.Task{
		{ID: 1, Title: "Pending Task 1"},
		{ID: 2, Title: "Pending Task 2"},
		{ID: 3, Title: "Completed Task 1", CompletedAt: &now},
		{ID: 4, Title: "Completed Task 2", CompletedAt: &now},
	}

	m := &Model{
		CoreModel: &CoreModel{
			tasks: tasks,
		},
	}

	testCases := []struct {
		name          string
		index         int
		showCompleted bool
		expectedID    int
		expectNil     bool
	}{
		{
			name:          "Get first pending task when completed are hidden",
			index:         0,
			showCompleted: false,
			expectedID:    1,
			expectNil:     false,
		},
		{
			name:          "Get second pending task when completed are hidden",
			index:         1,
			showCompleted: false,
			expectedID:    2,
			expectNil:     false,
		},
		{
			name:          "Index out of bounds for pending tasks",
			index:         2,
			showCompleted: false,
			expectNil:     true,
		},
		{
			name:          "Get first pending task when completed are shown",
			index:         0,
			showCompleted: true,
			expectedID:    1,
			expectNil:     false,
		},
		{
			name:          "Get first completed task when completed are shown",
			index:         2,
			showCompleted: true,
			expectedID:    3,
			expectNil:     false,
		},
		{
			name:          "Get second completed task when completed are shown",
			index:         3,
			showCompleted: true,
			expectedID:    4,
			expectNil:     false,
		},
		{
			name:          "Index out of bounds when completed are shown",
			index:         4,
			showCompleted: true,
			expectNil:     true,
		},
		{
			name:      "Negative index",
			index:     -1,
			expectNil: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m.showCompleted = tc.showCompleted
			task := m.getVisualTask(tc.index)

			if tc.expectNil {
				if task != nil {
					t.Errorf("Expected nil, but got task with ID %d", task.ID)
				}
			} else {
				if task == nil {
					t.Errorf("Expected task with ID %d, but got nil", tc.expectedID)
				} else if task.ID != tc.expectedID {
					t.Errorf("Expected task with ID %d, but got %d", tc.expectedID, task.ID)
				}
			}
		})
	}
}
