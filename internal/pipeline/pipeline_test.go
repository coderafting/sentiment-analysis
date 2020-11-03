package pipeline

import (
	"github.com/coderafting/sentiment-analysis/internal/database"
	"testing"
	"time"
)

func TestMemPartitions(t *testing.T) {
	type testCase struct {
		num        int
		buffer     int
		chansCount int
		bufferCap  int
	}
	cases := []testCase{
		testCase{num: 0, buffer: 0, chansCount: 1, bufferCap: 1},
		testCase{num: 0, buffer: 1, chansCount: 1, bufferCap: 1},
		testCase{num: 1, buffer: 0, chansCount: 1, bufferCap: 1},
		testCase{num: 2, buffer: 0, chansCount: 2, bufferCap: 1},
		testCase{num: 2, buffer: 2, chansCount: 2, bufferCap: 2},
	}
	for _, c := range cases {
		parts := MemPartitions(c.num, c.buffer)
		if c.chansCount != len(parts) && c.bufferCap != cap(parts[0]) {
			t.Errorf("Failed: expected chansCount: %v and bufferCap: %v, recieved %v and %v", c.chansCount, c.bufferCap, len(parts), cap(parts[0]))
		}
	}
}

func TestPubValidText(t *testing.T) {
	type testCase struct {
		validText  TweetText
		chansArray []chan TweetText
		indexRR    *MemRR
	}
	validText1 := TweetText{TextString: "I am happy"}
	validText2 := TweetText{TextString: "I am sad"}

	ch1 := make(chan TweetText, 1)
	ch2 := make(chan TweetText, 1)
	chans := []chan TweetText{ch1, ch2}

	rr := MemRR{Index: 0}

	cases := []testCase{
		{validText: validText1, chansArray: chans, indexRR: &rr},
		{validText: validText2, chansArray: chans, indexRR: &rr}}

	for _, c := range cases {
		PubValidText(c.validText, c.chansArray, c.indexRR)
		index := rr.Index
		out := <-c.chansArray[index]
		expected := c.validText.TextString
		outText := out.TextString
		if outText != expected {
			t.Errorf("Failed: expected %v, recieved %v", expected, outText)
		}
	}
}

func TestProcessText(t *testing.T) {
	testCase := TweetText{TextString: "I am happy and sad"}
	out := processText(testCase)
	if out[0].SentimentCategory != "jovility" && out[1].SentimentCategory != "sadness" {
		t.Errorf("Failed: recieved %v", out)
	}
}

func TestConsumeVTPubPT(t *testing.T) {
	validText := TweetText{TextString: "I am happy"}
	vtch := make(chan TweetText, 1)
	vtChans := []chan TweetText{vtch}

	outch1 := make(chan TweetText, 1)
	outchans := []chan TweetText{outch1}

	vtrr := MemRR{Index: 0}
	rr := MemRR{Index: 0}

	PubValidText(validText, vtChans, &vtrr)
	ConsumeVTPubPT(vtChans, outchans, &rr)
	out := <-outchans[rr.Index]
	if out.TextString != validText.TextString {
		t.Errorf("Failed: expected: %v, recieved %v", validText.TextString, out.TextString)
	}
}

func TestComputeAndSave(t *testing.T) {
	var mockDB = database.GetDatastore()
	validText := TweetText{TextString: "I am happy"}

	vtch := make(chan TweetText, 1)
	vtChans := []chan TweetText{vtch}
	outch1 := make(chan TweetText, 1)
	outchans := []chan TweetText{outch1}
	vtrr := MemRR{Index: 0}
	rr := MemRR{Index: 0}

	PubValidText(validText, vtChans, &vtrr)
	ConsumeVTPubPT(vtChans, outchans, &rr)
	ComputeAndSave(outchans, mockDB)
	time.Sleep(2 * time.Second) // just for a simplified testing
	out, _ := mockDB.FetchCategorySentiments("jovility")
	if out["jovility"].TextCount != 1 {
		t.Errorf("Failed: expected: %v, recieved %v", 1, out["jovility"].TextCount)
	}
}
