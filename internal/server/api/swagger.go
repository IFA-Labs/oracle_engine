package api

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag"
)

func (a *API) handleSwagger(w http.ResponseWriter, r *http.Request) {
	// If the request is for swagger.json, serve it directly
	if r.URL.Path == "/api/swagger.json" {
		w.Header().Set("Content-Type", "application/json")
		doc, err := swag.ReadDoc()
		if err != nil {
			http.Error(w, "Failed to read Swagger doc", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(doc))
		return
	}

	// For all other paths, use http-swagger to serve the UI
	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL("/api/swagger.json"), // The url pointing to API definition
	)
	swaggerHandler.ServeHTTP(w, r)
}
