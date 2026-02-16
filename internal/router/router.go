package router

import (
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"notification-system/docs"
	"notification-system/internal/config"
	"notification-system/internal/handler"
	"notification-system/internal/middleware"
	"notification-system/internal/queue"
	"notification-system/internal/repository"
	"notification-system/internal/service"
	"notification-system/internal/version"
)

// Deps holds dependencies required by the router.
type Deps struct {
	DB            *sqlx.DB
	UserRepo      repository.UserRepository
	MessageRepo   repository.MessageRepository
	RecipientRepo repository.RecipientRepository
	RedisClient   *redis.Client
	RateLimit     config.RateLimitConfig
	Publisher     *queue.Publisher
}

// NewRouter creates and configures the Gin engine with middleware and routes.
func NewRouter(deps Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Metrics())
	r.Use(middleware.CORS())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Version info
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, version.Info())
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Serve the raw OpenAPI spec from the embedded filesystem
	specFS, _ := fs.Sub(docs.SwaggerSpec, ".")
	r.StaticFS("/docs", http.FS(specFS))

	// Swagger UI — loads the spec from /docs/swagger.yaml
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/docs/swagger.yaml"),
	))

	// API v1 route group — protected by auth + rate limiting
	v1 := r.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(deps.UserRepo))
	v1.Use(middleware.RateLimitMiddleware(deps.RedisClient, deps.RateLimit))

	// Services
	msgService := service.NewMessageService(deps.DB, deps.MessageRepo, deps.RecipientRepo, deps.Publisher)

	// Message routes
	msgHandler := handler.NewMessageHandler(deps.DB, deps.MessageRepo, deps.RecipientRepo, msgService)
	messages := v1.Group("/messages")
	{
		messages.POST("/send", msgHandler.SendMessage)
		messages.POST("/bulk", msgHandler.BulkSend)
		messages.GET("/:id", msgHandler.GetMessageStatus)
		messages.GET("", msgHandler.ListMessages)
		messages.DELETE("/:id", msgHandler.CancelMessage)
	}

	// Webhook routes — unauthenticated (providers POST callbacks here)
	webhookHandler := handler.NewWebhookHandler(deps.RecipientRepo)
	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("/twilio", webhookHandler.TwilioWebhook)
		webhooks.POST("/sendgrid", webhookHandler.SendGridWebhook)
	}

	return r
}
