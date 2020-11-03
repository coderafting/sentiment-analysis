package database

/*
memorydb offers an implementation of the DataStore interface. It exposes an in-memory database.
*/

import (
	"fmt"
	"github.com/coderafting/panas-go/pkg/sentiment"
	"github.com/coderafting/sentiment-analysis/internal/utils"
	"log"
	"sync"
)

// MemoryDB indicates an in-memory database.
type MemoryDB struct {
	mux sync.Mutex
	db  Data
}

// GetDatastore instantiates a DataStore.
func GetDatastore() DataStore {
	db := MemoryDB{db: Data{Texts: map[ID]Text{}, Sentiments: map[Category]Sentiment{}, TotalTexts: 0}}
	return &db
}

// InsertText is exposed by DataStore interface. MemoryDB implements this method.
func (mdb *MemoryDB) InsertText(t string) (Text, error) {
	// No locking needed here, every insert is a unique insert
	id := ID(utils.GenerateUUID())
	txt := Text{ID: id, TextString: t}
	mdb.db.Texts[id] = txt
	inserted := mdb.db.Texts[id]
	if inserted != txt {
		return txt, fmt.Errorf("Failed to insert key: %v, with val: %v", id, t)
	}
	return txt, nil
}

// FetchSentiments returns the sentiment details of all the available categories.
func (mdb *MemoryDB) FetchSentiments() (map[Category]Sentiment, error) {
	return mdb.db.Sentiments, nil
}

// FetchCategorySentiments returns the Sentiment details of the supplied category.
func (mdb *MemoryDB) FetchCategorySentiments(catg string) (map[Category]Sentiment, error) {
	sents := mdb.db.Sentiments[Category(catg)]
	if sentiment.CategoriesMap[catg] != true {
		err := fmt.Errorf("Category %v doesn't exist", catg)
		return map[Category]Sentiment{Category(catg): sents}, err
	}
	return map[Category]Sentiment{Category(catg): sents}, nil
}

// UpdateSentiment updates the sentiment value of a supplied category as well as for other categories,
// based on the new texts count.
func (mdb *MemoryDB) UpdateSentiment(catg string, tcount int) (map[Category]Sentiment, error) {
	mdb.mux.Lock()
	catgSentiment := mdb.db.Sentiments[Category(catg)]
	// get the current counts
	oldCount := catgSentiment.TextCount
	oldTotal := mdb.db.TotalTexts
	// update the counts
	newCount := oldCount + tcount
	newTotal := oldTotal + tcount
	catgSentiment.TextCount = newCount
	mdb.db.TotalTexts = newTotal
	// update the sentiment of the category
	newSentimentVal, err := sentiment.CategoryAggregate(newCount, newTotal)
	if err != nil {
		log.Println("Error while computing aggregate sentiment.")
	}
	sentDetails := Sentiment{Value: newSentimentVal, TextCount: newCount}
	catgSentmnt := map[Category]Sentiment{Category(catg): sentDetails}
	mdb.db.Sentiments[Category(catg)] = sentDetails
	// Update all other sentiments based on the change in the total texts count
	allSents, err := mdb.FetchSentiments()
	if err != nil {
		log.Println("Error while fetching all sentiments, in the process of updating sentiments.")
	}
	for k, v := range allSents {
		if k != Category(catg) {
			newVal := (v.Value * float64(oldTotal)) / float64(newTotal)
			mdb.db.Sentiments[k] = Sentiment{Value: newVal, TextCount: v.TextCount}
		}
	}
	mdb.mux.Unlock()
	// Check if the update is successful, and return a response accordingly.
	updated := mdb.db.Sentiments[Category(catg)].Value
	if updated != newSentimentVal {
		return catgSentmnt, fmt.Errorf("Failed to update the sentiment of category: %v, with additional texts count of: %v", catg, tcount)
	}
	return catgSentmnt, nil
}
