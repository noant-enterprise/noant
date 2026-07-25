package domain

import "time"

type AuditLog struct {
	ID           string                 `json:"id" db:"id"`
	UserID       string                 `json:"user_id" db:"user_id"`
	UserName     string                 `json:"user_name" db:"user_name"`
	UserEmail    string                 `json:"user_email" db:"user_email"`
	OrgID        string                 `json:"org_id" db:"org_id"`
	Action       string                 `json:"action" db:"action"`
	ResourceType string                 `json:"resource_type" db:"resource_type"`
	ResourceID   *string                `json:"resource_id" db:"resource_id"`
	Details      map[string]interface{} `json:"details" db:"details"`
	IPAddress    *string                `json:"ip_address" db:"ip_address"`
	UserAgent    *string                `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}
