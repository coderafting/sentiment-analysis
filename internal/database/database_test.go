package database

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type StoreSuite struct {
	suite.Suite
	memDB MemoryDB
}

// Basic setup, instantiates the db
func (s *StoreSuite) SetupSuite() {
	s.memDB = MemoryDB{db: Data{Texts: map[ID]Text{}, Sentiments: map[Category]Sentiment{}, TotalTexts: 0}}
}

// Every test cycle will start with this state
func (s *StoreSuite) SetupTest() {
	s.memDB = MemoryDB{db: Data{Texts: map[ID]Text{}, Sentiments: map[Category]Sentiment{}, TotalTexts: 0}}
}

func TestStoreSuite(t *testing.T) {
	s := new(StoreSuite)
	suite.Run(t, s)
}

func (s *StoreSuite) TestInsertText() {
	testCase := "I am happy"
	txt, _ := s.memDB.InsertText(testCase)
	inserted := s.memDB.db.Texts[txt.ID].TextString
	expected := testCase
	if inserted != expected {
		s.T().Errorf("Insert failed, expected %v, got %v", expected, inserted)
	}
}

func (s *StoreSuite) TestFetchSentiments() {
	sentiments := map[Category]Sentiment{
		Category("jovility"): Sentiment{Value: 0.5, TextCount: 1},
		Category("sadness"):  Sentiment{Value: 0.5, TextCount: 1},
	}
	s.memDB.db.Sentiments = sentiments
	updatedSents := s.memDB.db.Sentiments
	for k, v := range updatedSents {
		if sentiments[k].Value != v.Value {
			s.T().Errorf("Fetch failed, expected %v, got %v", sentiments[k].Value, v.Value)
		}
	}
}

func (s *StoreSuite) TestFetchCategorySentiments() {
	sentiments := map[Category]Sentiment{
		Category("jovility"): Sentiment{Value: 0.5, TextCount: 1},
		Category("sadness"):  Sentiment{Value: 0.5, TextCount: 1},
	}
	s.memDB.db.Sentiments = sentiments
	testCategory := "jovility"
	updatedSents, _ := s.memDB.FetchCategorySentiments(testCategory)
	expected := sentiments[Category(testCategory)].Value
	out := updatedSents[Category(testCategory)].Value
	if expected != out {
		s.T().Errorf("Fetch failed, expected %v, got %v", expected, out)
	}
}

func (s *StoreSuite) TestUpdateSentiment() {
	s.memDB.db.Sentiments = map[Category]Sentiment{Category("jovility"): Sentiment{Value: 0.5, TextCount: 1}}
	s.memDB.db.TotalTexts = 1
	s.memDB.UpdateSentiment("sadness", 1)
	updatedJovSents := s.memDB.db.Sentiments[Category("jovility")].Value
	updatedJovTexts := s.memDB.db.Sentiments[Category("jovility")].TextCount
	updatedSadSents := s.memDB.db.Sentiments[Category("sadness")].Value
	updatedSadTexts := s.memDB.db.Sentiments[Category("sadness")].TextCount
	updatedTotalTexts := s.memDB.db.TotalTexts
	if updatedJovSents != 0.5 && updatedSadSents != 0.5 && updatedJovTexts != 1 && updatedSadTexts != 1 && updatedTotalTexts != 2 {
		s.T().Errorf("Update failed, updatedJovilitySents: %v, updatedJovilityTexts: %v, updatedSadnessSents: %v, updatedSadnessTexts: %v, updatedTotalTexts: %v", updatedJovSents, updatedJovTexts, updatedSadSents, updatedSadTexts, updatedTotalTexts)
	}
}
