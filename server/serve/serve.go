package serve

import (
	"fmt"
	"net/http"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"github.com/go-chi/chi/v5"
)

// deals with everything related to chi server

type ServeConfig struct {
	Port     string
	Schema   []interface{}
	Database *gorm.DB
	Router   *chi.Mux
}

func NewServeConfig(port string, schema []interface{}, database *gorm.DB, router *chi.Mux) *ServeConfig {
	return &ServeConfig{
		Port:     port,
		Schema:   schema,
		Database: database,
		Router:   router,
	}
}

func Serve(config *ServeConfig) {

	if config.Database == nil {
		panic("Database not initialized")
	}

	if config.Port == "" {
		config.Port = "3000"
	}

	// start chi server
	godotenv.Load()

	fmt.Println("CMS backend running on http://localhost:" + config.Port)

	fmt.Println("\nRoutes:\n")

	chi.Walk(config.Router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		fmt.Printf("[%s]: '%s' has %d middlewares\n", method, route, len(middlewares))
		return nil
	})

	http.ListenAndServe(":"+config.Port, config.Router)
}
