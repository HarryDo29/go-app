package dto

import "go-app/internal/schema"

type OverallStats struct {
	TotalUsers        int64           `json:"total_users"`
	ActiveUsers       int64           `json:"active_users"`
	BannedUsers       int64           `json:"banned_users"`
	NewThisWeek       int64           `json:"new_this_week"`
	LastRegistrations []schema.DbUser `json:"last_registrations"`
}
