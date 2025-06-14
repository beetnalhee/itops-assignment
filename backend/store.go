package main

import (
	"errors"
	"sync"
)

var (
	users = []*User{
		{ID: 1, Name: "김개발"},
		{ID: 2, Name: "이디자인"},
		{ID: 3, Name: "박기획"},
	}
	issues      = make(map[uint]*Issue)
	issuesMutex sync.RWMutex
	nextIssueID uint = 1
)

func GetUserByID(id uint) (*User, error) {
	for _, u := range users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, errors.New("존재하지 않는 사용자입니다.")
}

func AddIssue(issue *Issue) uint {
	issuesMutex.Lock()
	defer issuesMutex.Unlock()
	issue.ID = nextIssueID
	nextIssueID++
	issues[issue.ID] = issue
	return issue.ID
}

func GetIssueByID(id uint) (*Issue, error) {
	issuesMutex.RLock()
	defer issuesMutex.RUnlock()
	if issue, ok := issues[id]; ok {
		return issue, nil
	}
	return nil, errors.New("존재하지 않는 이슈입니다.")
}

func UpdateIssue(id uint, update func(*Issue) error) error {
	issuesMutex.Lock()
	defer issuesMutex.Unlock()
	issue, ok := issues[id]
	if !ok {
		return errors.New("존재하지 않는 이슈입니다.")
	}
	return update(issue)
}

func ListIssues(status string) []*Issue {
	issuesMutex.RLock()
	defer issuesMutex.RUnlock()
	var result []*Issue
	for _, issue := range issues {
		if status == "" || issue.Status == status {
			result = append(result, issue)
		}
	}
	return result
}
