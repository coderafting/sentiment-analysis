package service

import (
	"encoding/json"
	"github.com/coderafting/panas-go/pkg/sentiment"
	"github.com/coderafting/sentiment-analysis/config"
	"github.com/coderafting/sentiment-analysis/internal/database"
	"github.com/coderafting/sentiment-analysis/internal/pipeline"
	"github.com/coderafting/sentiment-analysis/internal/utils"
	"net/http"
)

// Handler exposes the base handler for the app.
type Handler struct {
	db                 database.DataStore
	cf                 config.Config
	validTextChans     []chan pipeline.TweetText
	processedTextChans []chan pipeline.TweetText
	vtRoundRobin       *pipeline.MemRR
	ptRoundRobin       *pipeline.MemRR
}

// GetHandler returns an instance of handler.
func GetHandler(db database.DataStore, c config.Config, vtc []chan pipeline.TweetText, ptc []chan pipeline.TweetText, vtr *pipeline.MemRR, ptr *pipeline.MemRR) *Handler {
	h := Handler{db, c, vtc, ptc, vtr, ptr}
	return &h
}

// SaveTextReq represents a textString key of type string, incoming via http request body.
type SaveTextReq struct {
	TextString string
}

// SaveTextResp is used for creating a response object for SaveText handler.
type SaveTextResp struct {
	Saved bool
}

// isValidText checks if the text is valid as per PANAS-t paper.
func isValidText(txt string) bool {
	return sentiment.ValidText(txt)
}

// SaveText is an http handler that saves the incoming text in DB,
// update the corresponding sentiment, and returns success object.
func (h *Handler) SaveText(w http.ResponseWriter, r *http.Request) {
	var tx SaveTextReq
	err := json.NewDecoder(r.Body).Decode(&tx)
	if err != nil {
		utils.JSONErrorResponse(w, err.Error())
		return
	}
	if tx.TextString == "" {
		utils.JSONErrorResponse(w, "Invalid body params")
		return
	}
	if isValidText(tx.TextString) {
		go pipeline.PubValidText(pipeline.TweetText{TextString: tx.TextString}, h.validTextChans, h.vtRoundRobin)
		data := SaveTextResp{Saved: true}
		utils.JSONSuccessResponse(w, data)
	} else {
		data := SaveTextResp{Saved: false}
		utils.JSONSuccessResponse(w, data)
	}
}

// GetSentiments is an http handler that returns all the sentiments, category-wise, from the db.
func (h *Handler) GetSentiments(w http.ResponseWriter, r *http.Request) {
	data, err := h.db.FetchSentiments()
	if err != nil {
		utils.JSONErrorResponse(w, err.Error())
		return
	}
	utils.JSONSuccessResponse(w, data)
}
