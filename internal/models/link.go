package models

import "time"

type Link struct {
    ID        string    `json:"id"`
    Code      string    `json:"code"`
    TargetURL string    `json:"target_url"`
    Clicks    int64     `json:"click_count"`
    CreatedAt time.Time `json:"created_at"`
    IsActive  bool      `json:"is_active"`
}
