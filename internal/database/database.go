// Package database exposes a DataStore interface.
package database

// ID represents ID of a Text, it is the string form of a UUID in the current implementation.
type ID string

// Category represents a sentiment category.
type Category string

// Text specifies the text object to be stored in the DB.
type Text struct {
	ID
	TextString string
}

// Sentiment represents the sentiment details of a category.
type Sentiment struct {
	Value     float64
	TextCount int
}

// Data is the main data-structure that the current implementation holds.
type Data struct {
	Texts      map[ID]Text
	Sentiments map[Category]Sentiment
	TotalTexts int
}

// DataStore is a database interface that can be implemented by different kinds of databases.
type DataStore interface {
	InsertText(t string) (Text, error)
	UpdateSentiment(catg string, tcount int) (map[Category]Sentiment, error)
	FetchSentiments() (map[Category]Sentiment, error)
	FetchCategorySentiments(catg string) (map[Category]Sentiment, error)
}
