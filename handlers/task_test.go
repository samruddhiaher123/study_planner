package handlers

import "testing"

func TestValidateTaskDefaultsAndLegacyPendingStatus(t *testing.T) {
	input := taskRequest{
		Title:     " Go study ",
		Subject:   " Backend ",
		Date:      "2026-10-12",
		StartTime: "09:00",
		EndTime:   "10:00",
		Status:    "Pending",
	}
	if message := validateTask(&input); message != "" {
		t.Fatalf("validateTask() = %q, want success", message)
	}
	if input.Status != "Not Started" {
		t.Errorf("status = %q, want Not Started", input.Status)
	}
	if input.Priority == nil || *input.Priority != "Medium" {
		t.Errorf("priority = %v, want Medium", input.Priority)
	}
	if input.Title != "Go study" || input.Subject != "Backend" {
		t.Errorf("task text was not trimmed: title %q, subject %q", input.Title, input.Subject)
	}
}

func TestValidateTaskRejectsInvalidStatusPriorityAndActualMinutes(t *testing.T) {
	tests := []struct {
		name   string
		update func(*taskRequest)
	}{
		{name: "status", update: func(input *taskRequest) { input.Status = "Unknown" }},
		{name: "priority", update: func(input *taskRequest) { input.Priority = stringPointer("Urgent") }},
		{name: "negative actual minutes", update: func(input *taskRequest) { input.ActualStudyMinutes = intPointer(-1) }},
		{name: "too many actual minutes", update: func(input *taskRequest) { input.ActualStudyMinutes = intPointer(1441) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validTaskRequest()
			test.update(&input)
			if message := validateTask(&input); message == "" {
				t.Fatal("validateTask() unexpectedly succeeded")
			}
		})
	}
}

func stringPointer(value string) *string { return &value }

func intPointer(value int) *int { return &value }

func validTaskRequest() taskRequest {
	return taskRequest{
		Title:     "Go study",
		Subject:   "Backend",
		Date:      "2026-10-12",
		StartTime: "09:00",
		EndTime:   "10:00",
	}
}
