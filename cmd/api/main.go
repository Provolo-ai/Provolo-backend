package main

import (
	"log"
	_ "provolo-api/docs"
	"provolo-api/internal/env"

	_ "github.com/joho/godotenv/autoload"
)

type application struct {
	port       int
	jwtSecret  string
	swaggerURL string
}

func main() {
	app := &application{
		port:       env.GetEnvInt("PORT", 8080),
		jwtSecret:  env.GetEnvString("JWT_SECRET", "secret"),
		swaggerURL: env.GetEnvString("SWAGGER_URL", "http://localhost:8080/swagger/doc.json"),
	}

	if err := app.serve(); err != nil {
		log.Fatal(err)
	}
}
