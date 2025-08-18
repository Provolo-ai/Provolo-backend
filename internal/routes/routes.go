package routes

import (
	"fmt"
	"log"
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

	// Environment-specific CORS
	if config.Environment == "production" {
		allowedOrigins := []string{
			"https://provolo.org",
			"https://www.provolo.org",
			"http://localhost:5173",
			"https://provolo-front-end-dev-env.vercel.app",
		}
		g.Use(middleware.CORSForProduction(allowedOrigins))
	} else {
		// Use development CORS (allows all origins)
		g.Use(middleware.CORS())
	}
	g.Use(middleware.Logger())

	// Initialize Firebase auth handler
	authHandler, err := handlers.NewAuthHandler()
	if err != nil {
		log.Fatalf("Failed to initialize auth handler: %v", err)
	}

	v1 := g.Group("/api/v1")
	{
		// Health
		v1.GET("/health", handlers.GetHealthCheck(*config))

		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.GET("/verify", authHandler.VerifySession)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes
		protected := v1.Group("/protected")
		protected.Use(middleware.AuthMiddleware(authHandler.GetClient()))
		{
			protected.GET("/profile", func(c *gin.Context) {
				userID := c.GetString("userID")
				userEmail := c.GetString("userEmail")
				displayName := c.GetString("displayName")
				c.JSON(http.StatusOK, types.NewSuccessResponse(
					"Profile Retrieved",
					"User profile data",
					gin.H{
						"userID":      userID,
						"email":       userEmail,
						"displayName": displayName,
					},
				))
			})
		}

		// Profile Optimization
		v1.Use(middleware.AuthMiddleware(authHandler.GetClient())).POST("/optimize-profile", handlers.ProfileOptimizer)

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
