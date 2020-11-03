package service

import (
	"bytes"
	"encoding/json"
	"github.com/coderafting/sentiment-analysis/config"
	"github.com/coderafting/sentiment-analysis/internal/database"
	"github.com/coderafting/sentiment-analysis/internal/pipeline"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIsValidText(t *testing.T) {
	type testCase struct {
		txt      string
		expected bool
	}
	cases := []testCase{
		testCase{txt: "I am happy", expected: true},
		testCase{txt: "It is a happy day", expected: false},
	}
	for _, c := range cases {
		out := isValidText(c.txt)
		if c.expected != out {
			t.Errorf("Failed: expected %v, got: %v", c.expected, out)
		}
	}
}

func TestSaveText(t *testing.T) {
	var mockDB = database.GetDatastore()
	var mockConfig = config.Config{Port: ":3000", Partitions: 4, PartitionBuffer: 10}
	var mockVTPIndex = pipeline.MemRR{Index: 0}
	var mockPTPIndex = pipeline.MemRR{Index: 0}
	var mockVTParts = pipeline.MemPartitions(4, 10)
	var mockPTParts = pipeline.MemPartitions(4, 10)
	var mockHandler = GetHandler(mockDB, mockConfig, mockVTParts, mockPTParts, &mockVTPIndex, &mockPTPIndex)

	reqData := SaveTextReq{TextString: "I am happy"}
	jsonReq, _ := json.Marshal(reqData)
	req, err := http.NewRequest("POST", "/text", bytes.NewBuffer(jsonReq))
	if err != nil {
		t.Fatal(err)
	}
	respRec := httptest.NewRecorder()
	handler := http.HandlerFunc(mockHandler.SaveText)
	handler.ServeHTTP(respRec, req)
	if status := respRec.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	time.Sleep(2 * time.Second) // just for a simplified testing
	expectedResp, _ := json.Marshal(SaveTextResp{Saved: true})
	expectedRespStr := bytes.NewBuffer(expectedResp).String()
	expected, _ := mockDB.FetchCategorySentiments("jovility")
	if respRec.Body.String() != expectedRespStr && expected["jovility"].TextCount != 0 {
		t.Errorf("handler returned unexpected body: got %v want %v", respRec.Body.String(), expected)
	}
}

func TestGetSentiments(t *testing.T) {
	var mockDB = database.GetDatastore()
	var mockConfig = config.Config{Port: ":3000", Partitions: 4, PartitionBuffer: 10}
	var mockVTPIndex = pipeline.MemRR{Index: 0}
	var mockPTPIndex = pipeline.MemRR{Index: 0}
	var mockVTParts = pipeline.MemPartitions(4, 10)
	var mockPTParts = pipeline.MemPartitions(4, 10)
	var mockHandler = GetHandler(mockDB, mockConfig, mockVTParts, mockPTParts, &mockVTPIndex, &mockPTPIndex)

	mockDB.UpdateSentiment("jovility", 1)
	req, err := http.NewRequest("GET", "/sentiments", nil)
	if err != nil {
		t.Fatal(err)
	}
	respRec := httptest.NewRecorder()
	handler := http.HandlerFunc(mockHandler.GetSentiments)
	handler.ServeHTTP(respRec, req)

	if status := respRec.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"jovility":{"Value":1,"TextCount":1}}`
	if respRec.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v, expected %v",
			respRec.Body.String(), expected)
	}
}
