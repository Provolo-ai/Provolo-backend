package types

// Default tier constants
const (
	DefaultTierID = "starter"
)

type Config struct {
	Port        int
	JwtSecret   string
	Environment string
	SwaggerURL  string
}

type Application struct {
	Config Config
}
