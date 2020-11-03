package service

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Routes specifies and returns the available http routes that are exposed as REST APIs.
// The allowed content type is JSON for all APIs.
// Authentication has been ignored for this demo service.
// A JWT based authentication can be implemented using the following packages:
// 		- github.com/dgrijalva/jwt-go
//		- github.com/go-chi/jwtauth
func Routes(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Post("/text", h.SaveText)
		r.Get("/sentiments", h.GetSentiments)
	})
	return r
}
