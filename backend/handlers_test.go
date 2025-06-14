package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAndGetIssue(t *testing.T) {
	r := NewRouter()
	// Create issue
	body := bytes.NewBufferString(`{"title":"테스트 이슈","description":"설명","userId":1}`)
	req := httptest.NewRequest("POST", "/issue", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var issue Issue
	if err := json.NewDecoder(w.Body).Decode(&issue); err != nil {
		t.Fatal(err)
	}
	if issue.Title != "테스트 이슈" || issue.Status != "IN_PROGRESS" || issue.User == nil || issue.User.ID != 1 {
		t.Fatalf("unexpected issue: %+v", issue)
	}
	// Get issue
	req2 := httptest.NewRequest("GET", "/issue/1", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	var got Issue
	if err := json.NewDecoder(w2.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.ID != 1 || got.Title != "테스트 이슈" {
		t.Fatalf("unexpected get: %+v", got)
	}
}

func TestPatchIssueStatusRules(t *testing.T) {
	r := NewRouter()
	// Create issue without user
	body := bytes.NewBufferString(`{"title":"상태테스트","description":"설명"}`)
	req := httptest.NewRequest("POST", "/issue", body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	// Try to patch to IN_PROGRESS without user
	patch := bytes.NewBufferString(`{"status":"IN_PROGRESS"}`)
	req2 := httptest.NewRequest("PATCH", "/issue/2", patch)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code == http.StatusOK {
		t.Fatal("should not allow IN_PROGRESS without user")
	}
	// Assign user, should auto change to IN_PROGRESS
	patch2 := bytes.NewBufferString(`{"userId":2}`)
	req3 := httptest.NewRequest("PATCH", "/issue/2", patch2)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w3.Code)
	}
	var updated Issue
	if err := json.NewDecoder(w3.Body).Decode(&updated); err != nil {
		t.Fatal(err)
	}
	if updated.Status != "IN_PROGRESS" || updated.User == nil || updated.User.ID != 2 {
		t.Fatalf("unexpected patch: %+v", updated)
	}
}
