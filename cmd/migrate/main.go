package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	migrateType := flag.String("type", "", "Type of migration to run (e.g., 'users_tier', 'prompt_quota')")
	flag.Parse()

	switch *migrateType {
	case "users_tier":
		migrateUsersTier()
	case "prompt_quota":
		migratePromptQuota()
	default:
		fmt.Println("Invalid migration type specified.")
		os.Exit(1)
	}
}
