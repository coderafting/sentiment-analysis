// Package service provides implementation for a community-sentiment-analysis server.
// It provides an in-memory database implementation, http handlers, http routes,
// and a mechanism to start the server on a given port.
// It also implements a two-stage data processing pipeline, which allows for parallel processing,
// exploiting the cores that the hardware has to offer.
package service

import (
	"github.com/coderafting/sentiment-analysis/config"
	"github.com/coderafting/sentiment-analysis/internal/database"
	"github.com/coderafting/sentiment-analysis/internal/pipeline"
	"github.com/go-chi/chi"
	"net/http"
)

// App contains the server configuration and http routes.
type App struct {
	Cf config.Config
	r  *chi.Mux
}

// There are two partition-sets (two sets of collection of channels) at the two stages of the pipeline,
// therefore, two round-robin states will have to be maintained.
var memVTPIndex = pipeline.MemRR{Index: 0}
var memPTPIndex = pipeline.MemRR{Index: 0}

// Initialization setup for an App instance.
func (a *App) init() {
	db := database.GetDatastore()
	validTextParts := pipeline.MemPartitions(a.Cf.Partitions, a.Cf.PartitionBuffer)
	processedTextParts := pipeline.MemPartitions(a.Cf.Partitions, a.Cf.PartitionBuffer)
	h := GetHandler(db, config.GetConfig(), validTextParts, processedTextParts, &memVTPIndex, &memPTPIndex)
	a.r = Routes(h)
	// initialize consumers and publishers for the 2nd and 3rd stage of the pipeline:
	// consumeVTPubPT fires goroutines that consume ValidTexts from a set of channels,
	// processe texts and publish to the next set of channels in the pipeline.
	go pipeline.ConsumeVTPubPT(h.validTextChans, h.processedTextChans, h.ptRoundRobin)
	// computeAndSave fires goroutines that consume ProcessedTexts from a set of channels,
	// and performs appropriate operations (save text and update sentiment) on the datastore.
	go pipeline.ComputeAndSave(h.processedTextChans, db)
}

// GetApp instantiates an app with its configuration, handlers, and routes.
func GetApp() *App {
	a := App{Cf: config.GetConfig()}
	a.init()
	return &a
}

// StartServer starts the http server for the supplied App instance.
func (a *App) StartServer() {
	http.ListenAndServe(a.Cf.Port, a.r)
}
