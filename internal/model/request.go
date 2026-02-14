package model

import "time"

// CreateMessageRequest is the API request body for sending a message.
type CreateMessageRequest struct {
	Subject     string     `json:"subject" binding:"required,max=200"`
	Message     string     `json:"message" binding:"required,max=5000"`
	From        string     `json:"from" binding:"required,max=100"`
	To          []string   `json:"to" binding:"required,min=1,max=1000,dive,required"`
	Platform    string     `json:"platform" binding:"required,oneof=sms whatsapp telegram email"`
	Priority    *int       `json:"priority,omitempty" binding:"omitempty,oneof=0 1 2"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

// BulkMessageRequest is the API request body for sending multiple messages.
type BulkMessageRequest struct {
	Messages []CreateMessageRequest `json:"messages" binding:"required,min=1,dive"`
}

// ListMessagesQuery represents the query parameters for listing messages.
type ListMessagesQuery struct {
	Page     int        `form:"page,default=1" binding:"min=1"`
	Limit    int        `form:"limit,default=20" binding:"min=1,max=100"`
	Platform string     `form:"platform" binding:"omitempty,oneof=sms whatsapp telegram email"`
	Status   *int       `form:"status" binding:"omitempty,min=0,max=6"`
	From     *time.Time `form:"from"`
	To       *time.Time `form:"to"`
}
