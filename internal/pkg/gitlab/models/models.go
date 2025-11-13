package models

type MergeRequest struct {
	ID           int    `json:"id"`
	IID          int    `json:"iid"`
	ProjectID    int    `json:"project_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	State        string `json:"state"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	TargetBranch string `json:"target_branch"`
	SourceBranch string `json:"source_branch"`
	Author       struct {
		Name string `json:"name"`
	} `json:"author"`
}
