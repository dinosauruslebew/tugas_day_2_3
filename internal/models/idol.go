package models

import "time"

type Idol struct {
    ID        int64      `json:"id"`
    Name      string     `json:"name"`
    Group     string     `json:"group"`
    Position  string     `json:"position"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    CreatedBy string     `json:"created_by"`
    UpdatedBy string     `json:"updated_by"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
    Version   int        `json:"version"`
}


