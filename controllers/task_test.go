package controllers

import (
	"net/http/httptest"
	"testing"
)

func TestParseTaskFilterNormalizesLegacyStatus(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/tasks?status=Pending&from=2026-10-01&to=2026-10-31&sort=priority", nil)

	filter, err := parseTaskFilter(request)
	if err != nil {
		t.Fatalf("parseTaskFilter() error = %v", err)
	}
	if filter.Status != "Not Started" {
		t.Errorf("filter status = %q, want Not Started", filter.Status)
	}
	if filter.From != "2026-10-01" || filter.To != "2026-10-31" || filter.Sort != "priority" {
		t.Errorf("parsed filter = %+v", filter)
	}
}

func TestParseTaskFilterRejectsInvalidDateRange(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/tasks?from=2026-11-01&to=2026-10-01", nil)
	if _, err := parseTaskFilter(request); err == nil || err.Error() != "from must be on or before to" {
		t.Errorf("parseTaskFilter() error = %v, want date range validation error", err)
	}
}

func TestValidateTaskAppliesExistingDefaults(t *testing.T) {
	input := taskRequest{
		Title:     " Revise biology ",
		Subject:   " Biology ",
		Date:      "2026-10-26",
		StartTime: "09:00",
		EndTime:   "10:00",
	}

	if message := validateTask(&input); message != "" {
		t.Fatalf("validateTask() = %q, want no error", message)
	}
	if input.Title != "Revise biology" || input.Subject != "Biology" {
		t.Errorf("trimmed title/subject = %q / %q", input.Title, input.Subject)
	}
	if input.Status != "Not Started" {
		t.Errorf("default status = %q, want Not Started", input.Status)
	}
	if input.Priority == nil || *input.Priority != "Medium" {
		t.Errorf("default priority = %v, want Medium", input.Priority)
	}
}
