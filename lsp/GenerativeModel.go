package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Message struct {
	sender  string
	content string
}

type ChatHistory struct {
}

type GenerativeModelAPI struct {
	endpoint string
	model    string
}

type CompletionResponse struct {
	Index   int    `json:"index"`
	Content string `json:"content"`
}

type CompletionRequest struct {
	Prompt string `json:"prompt"`
}

// Function: completion
// This function takes a prompt string and a Logger as parameters.
// The prompt string is used to generate a completion request and send it to the model API.
// The Logger is used to log debug messages and error messages.
//
// Parameters:
// - prompt: This is the input string which is used to create a CompletionRequest.
// - log: This is the logger used for debugging and error logging.
//
// Returns:
// - A string that is the completion of the prompt
func (m *GenerativeModelAPI) completion(prompt string, log commonlog.Logger) (string, error) {
	req := &CompletionRequest{Prompt: prompt}
	log.Debug(req.Prompt)
	json_data, err := json.Marshal(req)
	if err != nil {
		log.Error("Error encoding prompt")
		return "", err
	}
	log.Debug(string(json_data))
	resp, err := http.Post(
		m.endpoint+"completions",
		"application/string",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodybyte, err := io.ReadAll(resp.Body)
	var resp_data CompletionResponse
	json.Unmarshal(bodybyte, &resp_data)
	return resp_data.Content, err
}

func (GenerativeModelAPI) CreateTests(c string, file_type string) (string, error) {
	return "", nil
}

// Function: CreateComment
// This function generates a comment based on the provided code and file type.
//
// Parameters:
// - file_type (string): The type of file that the code is for.
// - code (string): The code snippet to generate a comment for.
// - log (Logger): The logger object to record events and errors.
//
// Returns:
// - (string, error): The generated comment or an error if one occurred.
func (m GenerativeModelAPI) CreateComment(file_type string, code string, log commonlog.Logger) (string, error) {
	_, dir, _, _ := runtime.Caller(0)
	dir = filepath.Dir(dir)
	prompt, err := os.ReadFile(filepath.Join(dir, "prompts/CreateComment.prmpt"))
	if err != nil {
		return "", fmt.Errorf("Could not read prmpts/CreateComment.prmpt from %s", dir)
	}
	prompt_str := fmt.Sprintf(string(prompt), file_type)
	resp, err := m.completion(prompt_str+code, log)
	if err != nil {
		return "", err
	}
	return FilterComments(resp), nil
}

func (GenerativeModelAPI) chat(c ChatHistory) {
}

func (GenerativeModelAPI) embeddings(prompt string) {
}

// RespToStr converts a CompletionResponse to a string.
//
// Parameters:
//   - resp: CompletionResponse to convert to a string.
//
// Returns:
//   - string: The string representation of the CompletionResponse's Content field.
func RespToStr(resp CompletionResponse) string {
	return resp.Content
}

// FilterComments function is used to remove non-comments from a given code string.
//
// It accepts a string parameter "code" which contains the source code.
// The function iterates through each line of the code and removes anything that
// isn't a comments.
//
// Returns:
// string: The code string with only comments
func FilterComments(code string) string {
	commentStyles := []string{"//", "#", "--"}
	blockCommentsOpener := []string{"/*", "\"\"\"", "'''", "--[["}
	blockCommentsCloser := map[string]string{
		"/*":     "*/",
		"/**":    "*/",
		"\"\"\"": "\"\"\"",
		"'''":    "'''",
		"--[[":   "]]",
	}
	lines := []string{}
	curOpener := ""
	for _, line := range strings.Split(code, "\n") {
		line_trimmed := strings.TrimSpace(line)
		if curOpener != "" {
			if strings.HasPrefix(line_trimmed, blockCommentsCloser[curOpener]) {
				curOpener = ""
			}
			lines = append(lines, line_trimmed)
		} else {
			for _, blockOpener := range blockCommentsOpener {
				if strings.HasPrefix(line_trimmed, blockOpener) {
					curOpener = blockOpener
					lines = append(lines, line_trimmed)
				}
			}
			for _, comment := range commentStyles {
				if strings.HasPrefix(line_trimmed, comment) {
					lines = append(lines, line_trimmed)
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}
