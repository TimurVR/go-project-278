// internal/dto/link.go
package dto

import (
    "regexp"
    "time"
    "net/url"
)

type Visit struct {
    Id        int       `json:"id" db:"id"`
    LinkID    int       `json:"link_id" db:"link_id"`
    IP        string    `json:"ip" db:"ip"`
    UserAgent string    `json:"user_agent" db:"user_agent"`
    Status    int       `json:"status" db:"status"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type LinkRequest struct {
    Original_url string `json:"original_url" binding:"required"`
    Short_name   string `json:"short_name,omitempty" binding:"omitempty,min=3,max=32"`
}


func (lr *LinkRequest) Validate() map[string]string {
    errors := make(map[string]string)
    if lr.Original_url == "" {
        errors["original_url"] = "обязательное поле"
    } else {
        u, err := url.ParseRequestURI(lr.Original_url)
        if err != nil || u.Scheme == "" || u.Host == "" {
            errors["original_url"] = "некорректный URL"
        }
    }
    if lr.Short_name != "" {
        if len(lr.Short_name) < 3 || len(lr.Short_name) > 32 {
            errors["short_name"] = "длина должна быть от 3 до 32 символов"
        }
        matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", lr.Short_name)
        if !matched {
            errors["short_name"] = "может содержать только буквы, цифры, дефисы и подчеркивания"
        }
    }
    return errors
}