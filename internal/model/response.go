package model

import "time"

// SendMessageResponse is returned after a message is successfully queued.
type SendMessageResponse struct {
	Success           bool      `json:"success"`
	MessageID         string    `json:"message_id"`
	RecipientsCount   int       `json:"recipients_count"`
	EstimatedDelivery time.Time `json:"estimated_delivery"`
	RequestID         string    `json:"request_id"`
}

// MessageStatusResponse is returned when querying the status of a message.
type MessageStatusResponse struct {
	Success   bool                `json:"success"`
	MessageID string              `json:"message_id"`
	Status    MessageStatusDetail `json:"status"`
}

// MessageStatusDetail contains the full status breakdown of a message.
type MessageStatusDetail struct {
	MessageID       string            `json:"message_id"`
	Subject         string            `json:"subject"`
	Platform        string            `json:"platform"`
	TotalRecipients int               `json:"total_recipients"`
	Summary         DeliverySummary   `json:"summary"`
	Recipients      []RecipientStatus `json:"recipients"`
	CreatedAt       time.Time         `json:"created_at"`
}

// DeliverySummary aggregates recipient status counts.
type DeliverySummary struct {
	Queued     int `json:"queued"`
	Processing int `json:"processing"`
	Sent       int `json:"sent"`
	Delivered  int `json:"delivered"`
	Failed     int `json:"failed"`
	Pending    int `json:"pending"`
}

// RecipientStatus is the per-recipient delivery status in a status response.
type RecipientStatus struct {
	Recipient   string     `json:"recipient"`
	Status      int        `json:"status"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
}

// ListMessagesResponse is the paginated list of messages.
type ListMessagesResponse struct {
	Success    bool       `json:"success"`
	Messages   []Message  `json:"messages"`
	Pagination Pagination `json:"pagination"`
}

// Pagination holds pagination metadata.
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ErrorResponse is the standard API error envelope.
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

// ErrorDetail describes the error.
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}
