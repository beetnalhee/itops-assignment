package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

var validStatuses = map[string]bool{
	"PENDING":     true,
	"IN_PROGRESS": true,
	"COMPLETED":   true,
	"CANCELLED":   true,
}

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/issue", issueHandler)
	mux.HandleFunc("/issues", issuesHandler)
	mux.HandleFunc("/issue/", issueDetailHandler)
	return mux
}

func issueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		createIssue(w, r)
		return
	}
	http.NotFound(w, r)
}

func issuesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		listIssues(w, r)
		return
	}
	http.NotFound(w, r)
}

func issueDetailHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/issue/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, "잘못된 이슈 ID입니다.", 400)
		return
	}
	if r.Method == http.MethodGet {
		getIssue(w, r, uint(id))
		return
	}
	if r.Method == http.MethodPatch {
		updateIssue(w, r, uint(id))
		return
	}
	http.NotFound(w, r)
}

func createIssue(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		UserID      *uint   `json:"userId"`
		Status      *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "잘못된 요청입니다.", 400)
		return
	}
	if req.Title == "" {
		writeError(w, "제목은 필수입니다.", 400)
		return
	}
	var user *User
	if req.UserID != nil {
		var err error
		user, err = GetUserByID(*req.UserID)
		if err != nil {
			writeError(w, err.Error(), 400)
			return
		}
	}
	// 상태 결정
	status := "PENDING"
	if user != nil {
		status = "IN_PROGRESS"
	}
	if req.Status != nil && validStatuses[*req.Status] {
		if (*req.Status == "IN_PROGRESS" || *req.Status == "COMPLETED" || *req.Status == "CANCELLED") && user == nil {
			writeError(w, "담당자가 없는 상태에서는 해당 상태로 변경할 수 없습니다.", 400)
			return
		}
		status = *req.Status
	}
	now := time.Now().UTC()
	issue := &Issue{
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		User:        user,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	AddIssue(issue)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(issue)
}

func listIssues(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status != "" && !validStatuses[status] {
		writeError(w, "유효하지 않은 상태입니다.", 400)
		return
	}
	issues := ListIssues(status)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"issues": issues})
}

func getIssue(w http.ResponseWriter, r *http.Request, id uint) {
	issue, err := GetIssueByID(id)
	if err != nil {
		writeError(w, err.Error(), 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issue)
}

func updateIssue(w http.ResponseWriter, r *http.Request, id uint) {
	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
		UserID      *uint   `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "잘못된 요청입니다.", 400)
		return
	}
	err := UpdateIssue(id, func(issue *Issue) error {
		if issue.Status == "COMPLETED" || issue.Status == "CANCELLED" {
			return errors.New("완료/취소된 이슈는 수정할 수 없습니다.")
		}
		if req.Title != nil {
			issue.Title = *req.Title
		}
		if req.Description != nil {
			issue.Description = *req.Description
		}
		if req.UserID != nil {
			if *req.UserID == 0 {
				issue.User = nil
				issue.Status = "PENDING"
			} else {
				user, err := GetUserByID(*req.UserID)
				if err != nil {
					return err
				}
				issue.User = user
				if issue.Status == "PENDING" && req.Status == nil {
					issue.Status = "IN_PROGRESS"
				}
			}
		}
		if req.Status != nil {
			if !validStatuses[*req.Status] {
				return errors.New("유효하지 않은 상태입니다.")
			}
			if (issue.User == nil) && (*req.Status == "IN_PROGRESS" || *req.Status == "COMPLETED") {
				return errors.New("담당자가 없는 상태에서는 해당 상태로 변경할 수 없습니다.")
			}
			issue.Status = *req.Status
		}
		// 상태가 COMPLETED 또는 CANCELLED로 바뀌면 이후 수정 불가
		if issue.Status == "COMPLETED" || issue.Status == "CANCELLED" {
			// 이후 PATCH에서 위에서 막히므로, 여기서는 단순히 상태만 반영
		}
		issue.UpdatedAt = time.Now().UTC()
		return nil
	})
	if err != nil {
		writeError(w, err.Error(), 400)
		return
	}
	issue, _ := GetIssueByID(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issue)
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg, Code: code})
}
