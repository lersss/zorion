package models

import "time"

type Contract struct {
    ID          string                 `json:"id"`
    SystemID    string                 `json:"system_id"`
    AuthorType  string                 `json:"author_type"`
    AuthorID    string                 `json:"author_id"`
    Title       string                 `json:"title"`
    Description string                 `json:"description"`
    ExpiresAt   time.Time              `json:"expires_at"`
    Status      string                 `json:"status"`
    Effects     map[string]interface{} `json:"effects"`
}