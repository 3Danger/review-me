package models

import (
	"strconv"
)

type Message struct {
	ServiceName  string
	MergeRequest MergeRequest
	JiraTask     JiraTask
}

func (m *Message) Short() string {
	return m.JiraTask.Short() + " " + m.MergeRequest.Short() + " - " + m.JiraTask.Summary
}

type MergeRequest struct {
	ID              int
	MergeRequestURL string
}

type JiraTask struct {
	ID      string
	Summary string
	Host    string
}

func (j *JiraTask) Short() string {
	return "[[" + j.ID + "](" + j.Host + "/" + j.ID + ")]"
}

func (m *MergeRequest) Short() string {
	return "[[MR-" + strconv.Itoa(m.ID) + "](" + m.MergeRequestURL + ")]"
}
