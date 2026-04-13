package models

type Job struct {
	ID       string `json:"id"`
	Task     string `json:"task"`
	Status   string `json:"status"`
	Retries  int    `json:"retries"`
	Priority int    `json:"priority"`
}
