package types

type Config struct {
	Port        int
	JwtSecret   string
	Environment string
	SwaggerURL  string
}

type Application struct {
	Config Config
}
