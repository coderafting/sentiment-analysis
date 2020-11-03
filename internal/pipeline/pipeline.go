// Package pipeline provides implementation for a multi-stage sentiment analysis pipeline.
// The analysis is based on the PANAS-t paper, using Twitter data.
// It also provides a mechanism to perform parallel processing in the pipeline, and a mechanism
// to load-balance while achieving parallelism.
package pipeline

import (
	"github.com/coderafting/panas-go/pkg/sentiment"
	"github.com/coderafting/sentiment-analysis/internal/database"
	"log"
	"sync"
)

// TweetText represents a valid text string that can be considered for sentiment analysis,
// based on the criteria set by the PANAS-t paper.
type TweetText struct {
	TextString        string
	SentimentCategory string
}

/*
Create a collection of channels (partitions) that can be used at different stages of the pipeline
to enable the parallel processing.
Also, provide a mechanism to balance the load on the channels in a partition.
RoundRobin mechanism has been implemented to achieve load-balancing on the partitions in the pipeline.
*/

// MemPartitions returns a slice of buffered channels, based on the partition count and buffer value.
func MemPartitions(num int, buffer int) []chan TweetText {
	var partitions []chan TweetText
	bufferCount := buffer
	if buffer <= 0 {
		bufferCount = 1
	}
	if num <= 0 {
		partitions = append(partitions, make(chan TweetText, bufferCount))
	} else {
		for n := 1; n <= num; n++ {
			partitions = append(partitions, make(chan TweetText, bufferCount))
		}
	}
	return partitions
}

// MemRR represents an in-memory RoundRobin object.
type MemRR struct {
	mux   sync.Mutex
	Index int
}

// NextIndex consumes the in-memory RoundRobin object,
// increases its Index value by 1, if it is less than the specified max value, and returns the Index value.
// If the Index reaches the max, it resets it to 0, and returns the Index.
func NextIndex(m *MemRR, max int) int {
	res := 0
	m.mux.Lock()
	if m.Index == max {
		m.Index = 0
	} else {
		m.Index++
		res++
	}
	m.mux.Unlock()
	return res
}

// PubValidText publishes valid TweetText data to an appropriate partition (channel) from a collection of channels.
// The selection of a channel happens via round-robin mechanism.
func PubValidText(vt TweetText, chansArray []chan TweetText, indexRR *MemRR) {
	index := NextIndex(indexRR, len(chansArray)-1)
	chansArray[index] <- vt
}

// processText transforms TweetText into a collection of processed TweetText data with their Sentiment Categories.
func processText(txt TweetText) []TweetText {
	ctgs := sentiment.Categories(txt.TextString)
	pts := []TweetText{}
	for _, c := range ctgs {
		pts = append(pts, TweetText{TextString: txt.TextString, SentimentCategory: c})
	}
	return pts
}

/*
Operators that will consume from a channel, transform data, and publish to a channel at different
stages of the pipeline.
*/

// PubProcessedText consumes from a channel of TweetText, process the data, and publishes to
// an appropriate channel from a collection of channels.
// The selection of a channel happens via round-robin mechanism.
func PubProcessedText(in chan TweetText, out []chan TweetText, indexRR *MemRR) {
	index := NextIndex(indexRR, len(out)-1)
	for vt := range in {
		// Note that the following for loop is alright because
		// the result of processText(vt) will be often a slice of 1 or 2 items.
		// The length of the resultant slice depends upon the number of sentiment categories
		// a text contains – this is mostly 1, and in some cases 2.
		// It is rare that a sentence will carry more than 2 sentiment categories.
		for _, t := range processText(vt) {
			out[index] <- t
		}
	}
}

// ConsumeVTPubPT triggers goroutines, each of which starts consuming a specific channel,
// and produce the processed data to a channel from a collection of channels.
func ConsumeVTPubPT(in []chan TweetText, out []chan TweetText, indexRR *MemRR) {
	for _, c := range in {
		go PubProcessedText(c, out, indexRR)
	}
}

// ComputeSentimentAndSave consumes processed TweetText from a channel, saves the text to DB, and
// updates the corresponding sentiments.
func ComputeSentimentAndSave(in chan TweetText, db database.DataStore) {
	for pt := range in {
		txt := pt.TextString
		ctg := pt.SentimentCategory
		// Save text
		_, insertErr := db.InsertText(txt)
		if insertErr != nil {
			log.Printf("Error inserting the text: %v", txt)
		}
		// Update sentiment
		_, updateErr := db.UpdateSentiment(ctg, 1)
		if updateErr != nil {
			log.Printf("Error updating sentiment for category: %v, with text: %v", ctg, txt)
		}
	}
}

// ComputeAndSave triggers goroutines, each of which starts consuming a specific channel,
// and perform computeSentimentAndSave operation on DB.
func ComputeAndSave(in []chan TweetText, db database.DataStore) {
	for _, c := range in {
		go ComputeSentimentAndSave(c, db)
	}
}
