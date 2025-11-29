package main

import (
	"testing"
	"time"

	"github.com/tliron/commonlog"
)

func NewModelAPI() GenerativeModelAPI {
	return GenerativeModelAPI{
		endpoint: "http://localhost:8000/",
		model:    "codeCompletions",
	}
}
func NewTestCode() []string {
	var result []string
	result = append(result, "Hello")
	result = append(result, "World")
	result = append(result, "Hello")
	result = append(result, "World")
	return result
}

func TestCompletionsAPI(t *testing.T) {
	ConfigureLogger()
	log = commonlog.GetLogger("Test")
	model := NewModelAPI()
	start_time := time.Now()
	result, err := model.CodeCompletion(NewTestCode(), 2, 3, log)
	if err != nil {
		t.Error(err.Error())
	}
	duration := time.Since(start_time)
	log.Infof("RESULT::: %s", result)
	log.Infof("DURATION::: %s", duration)
	time.Sleep(1000)
	if result == "" {
		t.Error("No completion generated!")
	}
}
