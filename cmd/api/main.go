package main

import (
	"fmt"
	"log"
	_ "provolo-api/docs"
	"provolo-api/internal/env"
	"provolo-api/internal/routes"
	"provolo-api/internal/types"

	_ "github.com/joho/godotenv/autoload"
)

// @title Provolo API
// @version 1.0
// @description This is the Provolo backend API server
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host provolo-api.onrender.com
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	port := env.GetEnvInt("PORT", 8080)
	host := env.GetEnvString("HOST", "localhost")

	config := &types.Config{
		Port:        port,
		JwtSecret:   env.GetEnvString("JWT_SECRET", "secret"),
		Environment: env.GetEnvString("ENVIRONMENT", "development"),
		SwaggerURL:  env.GetEnvString("SWAGGER_URL", fmt.Sprintf("http://%s:%d/swagger/doc.json", host, port)),
	}

	if err := routes.StartServer(config); err != nil {
		log.Fatal(err)
	}
}
