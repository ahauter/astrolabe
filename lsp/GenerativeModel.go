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

func (m *GenerativeModelAPI) completion(prompt string, log commonlog.Logger) (string, error) {
	req := &CompletionRequest{Prompt: prompt + "\n<|im_start|>assistant\n"}
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

func wrapPrompt(prompt string, role string) string {
	result := fmt.Sprintf(`<|im_start|>%s %s<|im_end|>`, role, prompt)
	return result
}

func (m GenerativeModelAPI) CreateComment(file_type string, code string, log commonlog.Logger) (string, error) {
	_, dir, _, _ := runtime.Caller(0)
	dir = filepath.Dir(dir)
	prompt, err := os.ReadFile(filepath.Join(dir, "prompts/CreateComment.prmpt"))
	if err != nil {
		return "", fmt.Errorf("Could not read prmpts/CreateComment.prmpt from %s", dir)
	}
	prompt_str := fmt.Sprintf(string(prompt), file_type)
	code = wrapPrompt(code, "system")
	prompt_str = wrapPrompt(prompt_str, "system")
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

func RespToStr(resp CompletionResponse) string {

	return resp.Content
}

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
