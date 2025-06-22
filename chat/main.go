package main

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gap = "\n\n"

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type ResponseFormat struct {
	Type   string `json:"type"`
	Strict bool   `json:"strict"`
	Schema Schema `json:"schema"`
}

type properties struct {
	Response struct {
		Type string `json:"type"`
	} `json:"response"`
}

type Schema struct {
	Type                 string     `json:"type"`
	Properties           properties `json:"properties"`
	Required             []string   `json:"required"`
	AdditionalProperties bool       `json:"additionalProperties"`
}

func makeSchema() Schema {
	return Schema{
		Type: "object",
		Properties: properties{
			Response: struct {
				Type string "json:\"type\""
			}{
				Type: "string",
			},
		},
		Required:             []string{"response"},
		AdditionalProperties: false,
	}
}
func makeDefaultResponseFormat() ResponseFormat {
	return ResponseFormat{
		Type:   "json_object",
		Strict: true,
		Schema: makeSchema(),
	}
}

type messageRequest struct {
	Messages       []message      `json:"messages"`
	ResponseFormat ResponseFormat `json:"response_format"`
}

func makeDefaultMessageRequest(messages []message) messageRequest {
	return messageRequest{
		Messages:       messages,
		ResponseFormat: makeDefaultResponseFormat(),
	}
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type model struct {
	viewport    viewport.Model
	messages    []message
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
}

func (m message) toString() string {
	return fmt.Sprintf("%s: %s", m.Role, m.Content)
}

func makeMessage(sender string, value string) message {
	return message{
		sender,
		value,
	}
}

func makeMessageTexts(messages []message) string {
	messageTexts := []string{}
	for i := 0; i < len(messages); i++ {
		messageTexts = append(messageTexts, messages[i].toString())
	}
	return strings.Join(messageTexts, "\n")
}

type ResponseCallback func(resp *http.Response, err error)

const endpoint = "http://127.0.0.1:9999/v1/chat/completions"

func sendChatRequest(
	ctx context.Context,
	messages []message,
	callback ResponseCallback,
) context.CancelFunc {
	ctx, can := context.WithCancel(ctx)
	//format the messages
	message_obj := makeDefaultMessageRequest(messages)
	message_json, err := json.Marshal(message_obj)
	if err != nil {
		fmt.Println("Error formatting message json")
		message_json = []byte("[]")
	}
	req, err := http.NewRequestWithContext(
		ctx, "POST", endpoint, bytes.NewBuffer(message_json),
	)
	if err != nil {
		fmt.Println("Error create request!")
		return can
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		callback(nil, err)
	} else {
		callback(resp, nil)
	}
	return can
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		messages:    []message{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

var cancel context.CancelFunc = nil

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		if len(m.messages) > 0 {
			// Wrap content before setting it.
			m.viewport.SetContent(
				lipgloss.NewStyle().Width(m.viewport.Width).Render(
					makeMessageTexts(m.messages)),
			)
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.messages = append(m.messages, makeMessage("User", m.textarea.Value()))
			callback := func(resp *http.Response, err error) {
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				fmt.Println(string(body))
				var chat_completion map[string]interface{}
				err = json.Unmarshal(body, &chat_completion)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				assistant_choice, ok := chat_completion["choices"].([]interface{})
				if !ok {
					fmt.Println("Cant read choices")
				}
				choices_raw, ok := assistant_choice[0].(map[string]interface{})
				if !ok {
					fmt.Println("Cant read 0")
				}
				assistant_message_info, ok := choices_raw["message"].(map[string]interface{})
				if !ok {
					fmt.Println("Cant read message")
				}
				assistant_message, ok := assistant_message_info["content"].(string)
				if !ok {
					fmt.Println("Cant read content")
				}
				var message_info map[string]interface{}
				json.Unmarshal([]byte(assistant_message), &message_info)
				message, ok := message_info["response"].(string)
				if !ok {
					fmt.Println("Cant read response")
				}
				m.messages = append(m.messages, makeMessage("Assistant", message))
				m.viewport.SetContent(
					lipgloss.NewStyle().Width(m.viewport.Width).Render(makeMessageTexts(m.messages)),
				)
				m.textarea.Reset()
				m.viewport.GotoBottom()
			}
			// cancel any current requests
			if cancel != nil {
				cancel()
				cancel = nil
			}
			// send a new chat request to the model
			cancel = sendChatRequest(
				context.Background(),
				m.messages,
				callback,
			)

			m.viewport.SetContent(
				lipgloss.NewStyle().Width(m.viewport.Width).Render(makeMessageTexts(m.messages)),
			)
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}
