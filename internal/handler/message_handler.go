package handler

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"notification-system/internal/middleware"
	"notification-system/internal/model"
	"notification-system/internal/repository"
	"notification-system/internal/service"
	"notification-system/pkg/logger"
)

// MessageHandler handles HTTP requests for messages.
type MessageHandler struct {
	db            *sqlx.DB
	messageRepo   repository.MessageRepository
	recipientRepo repository.RecipientRepository
	service       *service.MessageService
}

// NewMessageHandler creates a new MessageHandler.
func NewMessageHandler(
	db *sqlx.DB,
	messageRepo repository.MessageRepository,
	recipientRepo repository.RecipientRepository,
	service *service.MessageService,
) *MessageHandler {
	return &MessageHandler{
		db:            db,
		messageRepo:   messageRepo,
		recipientRepo: recipientRepo,
		service:       service,
	}
}

// SendMessage handles POST /api/v1/messages/send
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req model.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error: model.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "UNAUTHORIZED", Message: "User not found in context"},
		})
		return
	}

	resp, err := h.service.SendMessage(c.Request.Context(), user.ID, req)
	if err != nil {
		logger.Get().Error().Err(err).Msg("failed to process message request")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Failed to process request"},
		})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetMessageStatus handles GET /api/v1/messages/:id
func (h *MessageHandler) GetMessageStatus(c *gin.Context) {
	idStr := c.Param("id")
	msgID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "VALIDATION_ERROR", Message: "Invalid message ID format"},
		})
		return
	}

	msg, err := h.messageRepo.GetByID(c.Request.Context(), msgID)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, model.ErrorResponse{
				Success: false,
				Error:   model.ErrorDetail{Code: "NOT_FOUND", Message: "Message not found"},
			})
			return
		}
		logger.Get().Error().Err(err).Msg("failed to get message")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Failed to get message"},
		})
		return
	}

	recipients, err := h.recipientRepo.GetByMessageID(c.Request.Context(), msgID)
	if err != nil {
		logger.Get().Error().Err(err).Msg("failed to get recipients")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Failed to get recipients"},
		})
		return
	}

	// Build summary
	var summary model.DeliverySummary
	recipientStatuses := make([]model.RecipientStatus, len(recipients))
	for i, r := range recipients {
		switch r.Status {
		case model.StatusQueued:
			summary.Queued++
		case model.StatusProcessing:
			summary.Processing++
		case model.StatusSent:
			summary.Sent++
		case model.StatusDelivered:
			summary.Delivered++
		case model.StatusFailed:
			summary.Failed++
		case model.StatusPending:
			summary.Pending++
		}

		recipientStatuses[i] = model.RecipientStatus{
			Recipient:   r.Recipient,
			Status:      int(r.Status),
			SentAt:      r.SentAt,
			DeliveredAt: r.DeliveredAt,
		}
	}

	c.JSON(http.StatusOK, model.MessageStatusResponse{
		Success:   true,
		MessageID: msg.ID.String(),
		Status: model.MessageStatusDetail{
			MessageID:       msg.ID.String(),
			Subject:         msg.Subject,
			Platform:        string(msg.Platform),
			TotalRecipients: len(recipients),
			Summary:         summary,
			Recipients:      recipientStatuses,
			CreatedAt:       msg.CreatedAt,
		},
	})
}

// ListMessages handles GET /api/v1/messages
func (h *MessageHandler) ListMessages(c *gin.Context) {
	var query model.ListMessagesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	// Defaults
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "UNAUTHORIZED", Message: "User not found in context"},
		})
		return
	}

	messages, total, err := h.messageRepo.List(c.Request.Context(), user.ID, query)
	if err != nil {
		logger.Get().Error().Err(err).Msg("failed to list messages")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Failed to list messages"},
		})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(query.Limit)))

	c.JSON(http.StatusOK, model.ListMessagesResponse{
		Success:  true,
		Messages: messages,
		Pagination: model.Pagination{
			Page:       query.Page,
			Limit:      query.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// BulkSend handles POST /api/v1/messages/bulk
func (h *MessageHandler) BulkSend(c *gin.Context) {
	var req model.BulkMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	user := middleware.GetUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "UNAUTHORIZED", Message: "User not found in context"},
		})
		return
	}

	var results []model.BulkMessageResult
	for i, msg := range req.Messages {
		resp, err := h.service.SendMessage(c.Request.Context(), user.ID, msg)
		if err != nil {
			logger.Get().Error().Err(err).Int("index", i).Msg("bulk: failed to send message")
			results = append(results, model.BulkMessageResult{
				Index:   i,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}
		results = append(results, model.BulkMessageResult{
			Index:     i,
			Success:   true,
			MessageID: resp.MessageID,
		})
	}

	c.JSON(http.StatusCreated, model.BulkMessageResponse{
		Success:    true,
		Total:      len(req.Messages),
		Successful: countSuccessful(results),
		Failed:     len(req.Messages) - countSuccessful(results),
		Results:    results,
	})
}

// CancelMessage handles DELETE /api/v1/messages/:id
func (h *MessageHandler) CancelMessage(c *gin.Context) {
	idStr := c.Param("id")
	msgID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "VALIDATION_ERROR", Message: "Invalid message ID format"},
		})
		return
	}

	msg, err := h.messageRepo.GetByID(c.Request.Context(), msgID)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, model.ErrorResponse{
				Success: false,
				Error:   model.ErrorDetail{Code: "NOT_FOUND", Message: "Message not found"},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Failed to get message"},
		})
		return
	}

	// Only scheduled messages can be cancelled
	if msg.Status != model.StatusScheduled {
		c.JSON(http.StatusConflict, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "INVALID_STATE", Message: "Only scheduled messages can be cancelled"},
		})
		return
	}

	if err := h.messageRepo.UpdateStatus(c.Request.Context(), msgID, model.StatusCancelled); err != nil {
		logger.Get().Error().Err(err).Msg("failed to cancel message")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   model.ErrorDetail{Code: "INTERNAL_ERROR", Message: "Failed to cancel message"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message_id": msgID.String(),
		"status":     "cancelled",
	})
}

func countSuccessful(results []model.BulkMessageResult) int {
	count := 0
	for _, r := range results {
		if r.Success {
			count++
		}
	}
	return count
}
