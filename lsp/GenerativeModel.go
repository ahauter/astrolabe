package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

func (m *GenerativeModelAPI) completion(prompt string) (string, error) {
	req := &CompletionRequest{Prompt: prompt + "\n<|im_start|>assistant\n"}
	log.Println(req.Prompt)
	json_data, err := json.Marshal(req)
	if err != nil {
		log.Println("Error encoding prompt")
		return "", err
	}
	log.Println(string(json_data))
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

func (m GenerativeModelAPI) CreateComment(file_type string, code string) (string, error) {
	prompt := fmt.Sprintf(`You are a technical writer documenting a codebase.
		Write a specific, detailed comment for the following code,
		include every detail you can find about parameters, return, and possible errors.
		You MUST Only write the function header comment.
		You will be penalized if you write any code.
		You will be penalized if you write an imcomplete comment
		The file type is %s.
		Use the following comment symbols appropriate for the file type: #, //, or --
		\n`,
		file_type,
	)
	code = wrapPrompt(code, "system")
	prompt = wrapPrompt(prompt, "system")
	resp, err := m.completion(prompt + code)
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
	//TODO multiline comments
	lines := []string{}
	for _, line := range strings.Split(code, "\n") {
		for _, comment := range commentStyles {
			line_trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(line_trimmed, comment) {
				lines = append(lines, line_trimmed)
			}
		}
	}
	return strings.Join(lines, "\n")
}
