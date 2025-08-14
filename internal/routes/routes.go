package routes

import (
	"fmt"
	"net/http"
	"provolo-api/internal/handlers"
	"provolo-api/internal/middleware"
	"provolo-api/internal/types"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all application routes
func SetupRoutes(config *types.Config) http.Handler {
	g := gin.Default()

	// Global middleware
	g.Use(middleware.CORS())
	g.Use(middleware.Logger())

	// Health check (before API group for easier monitoring)
	g.GET("/health", handlers.GetHealthCheck(*config))

	v1 := g.Group("/api/v1")
	{
		// Health (also available in API group)
		v1.GET("/health", handlers.GetHealthCheck(*config))

		// Payments Webhook
		v1.POST("/payment-webhook", handlers.PaymentWebhook)
	}

	// Swagger documentation with dynamic URL
	swaggerURL := fmt.Sprintf("http://localhost:%d/swagger/doc.json", config.Port)
	if config.Environment == "production" {
		swaggerURL = config.SwaggerURL
	}

	g.GET("/swagger/*any", func(c *gin.Context) {
		if c.Request.RequestURI == "/swagger/" {
			c.Redirect(302, "/swagger/index.html")
		}
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL(swaggerURL))(c)
	})

	return g
}
