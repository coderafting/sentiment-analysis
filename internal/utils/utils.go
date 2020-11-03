// Package utils exposes some utility functions that can be used by other packages.
package utils

import (
	"github.com/google/uuid"
	"github.com/unrolled/render"
	"net/http"
)

// GenerateUUID generates a unique UUID in the string format.
func GenerateUUID() string {
	return uuid.New().String()
}

// JSONSuccessResponse creates an http success response object.
func JSONSuccessResponse(w http.ResponseWriter, data interface{}) {
	re := render.New()
	re.JSON(w, http.StatusOK, data)
}

// JSONNoDataResponse creates an http no-data response object.
func JSONNoDataResponse(w http.ResponseWriter, respStr string) {
	re := render.New()
	re.JSON(w, http.StatusNoContent, respStr)
}

// JSONErrorResponse creates an http error response object.
func JSONErrorResponse(w http.ResponseWriter, err string) {
	re := render.New()
	re.JSON(w, http.StatusBadRequest, err)
}
