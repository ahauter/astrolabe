package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
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

type CompletionItem struct {
	Index   int    `json:"index"`
	Content string `json:"text"`
}

type CompletionResponse struct {
	Choices []CompletionItem `json:"choices"`
}

type CompletionRequest struct {
	Prompt string `json:"prompt"`
}

/*
Completes the given prompt using the Generative Model API.

Parameters:
- prompt: The prompt to complete.
- log: The logger to use for logging.

Returns:
- string: The completed prompt.
- error: Any error that occurred during the completion.

Errors:
- Returns an error if there was an issue marshaling the request to JSON.
- Returns an error if there was an issue making the HTTP request.
- Returns an error if there was an issue unmarshaling the response from JSON.
*/
func (m *GenerativeModelAPI) completion(prompt string, log commonlog.Logger) (string, error) {
	req := &CompletionRequest{Prompt: prompt}
	//log.Debug(req.Prompt)
	json_data, err := json.Marshal(req)
	if err != nil {
		log.Error("Error encoding prompt")
		return "", err
	}
	//log.Debug(string(json_data))
	resp, err := http.Post(
		m.endpoint+"v1/completions",
		"application/json",
		bytes.NewBuffer(json_data),
	)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		bodybyte, err := io.ReadAll(resp.Body)
		var resp_data CompletionResponse
		json.Unmarshal(bodybyte, &resp_data)
		if len(resp_data.Choices) > 0 {
			return resp_data.Choices[0].Content, err
		}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warningf("Got error reading body of response in error!\n%s", err.Error())
	}
	log.Errorf("Error in response: %s", string(body))
	return "", err
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

const CURSOR_TOKEN = "<CURSOR>"

func (m GenerativeModelAPI) CodeCompletion(code []string, line int, character int, log commonlog.Logger) (string, error) {
	log.Debugf("Completions request at %s, %s", line, character)
	//log.Debugf("Completions lines, %s", strings.Join(code, "\n"))
	if line >= len(code) || line < 0 {
		log.Warning("Line is outside document bounds!")
		return "", nil
	}
	target_line := code[line]
	log.Debugf("Target line: ", target_line)
	full_prompt := []string{}
	new_target := target_line + CURSOR_TOKEN
	if character < len(target_line) && character >= 0 {
		new_target = target_line[:character] + CURSOR_TOKEN + target_line[character:]
	} else {
		log.Warning("Character is outside line bounds!")
	}
	code[line] = new_target
	full_prompt = append(full_prompt, code...)
	full_prompt = append(full_prompt, "ENDOFFILE")
	full_prompt = append(full_prompt, code[:line]...)
	prompt := strings.Join(full_prompt, "\n")
	resp, err := m.completion(prompt, log)
	if err != nil {
		return "", err
	}
	// only return first line
	lines := strings.Split(resp, "\n")
	result := ""
	for _, line := range lines {
		if len(strings.TrimSpace(line)) != 0 {
			result = line
			break
		}
	}
	return result, nil
}

func (GenerativeModelAPI) chat(c ChatHistory) {
}

func (GenerativeModelAPI) embeddings(prompt string) {
}

// RespToStr converts a CompletionResponse to a string.
//
// Parameters:{
//   - resp: CompletionResponse to convert to a string.
//
// Returns:
//   - string: The string representation of the CompletionResponse's Content field.
func RespToStr(resp CompletionResponse) string {
	if 0 < len(resp.Choices) {
		return resp.Choices[0].Content
	}
	return ""
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
