package gemini

import (
	"fmt"
	"log"
	"os"
	"strings"

	"context"

	genai "github.com/google/generative-ai-go/genai"
	dotEnv "github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	fmt.Println("Hello world")
	model, err := NewGenerationModel()
	if err != nil {
		log.Fatalf(err.Error())
	}
	comment, err := model.CreateComment(
		` 
func FilterComments(code string) string {
	commentStyles := []string{"//", "#"}
	//TODO multiline comments
	lines := []string{}
	for _, line := range strings.Split(code, "\n") {
		for _, comment := range commentStyles {
			line_trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(line_trimmed, comment) {
				lines = append(lines, line)
			}
		}
	}
	return strings.Join(lines, "\n")
}

    `,
	)

	fmt.Println(model.CreateTests(comment))
}

type GenerationModel struct {
	context context.Context
	model   *genai.GenerativeModel
}

func NewGenerationModel() (*GenerationModel, error) {
	err := dotEnv.Load(".env")
	if err != nil {
		return nil, err
	}
	api_key := os.Getenv("API_KEY")
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(api_key))
	if err != nil {
		return nil, err
	}
	model := client.GenerativeModel("gemini-pro")
	return &GenerationModel{context: ctx, model: model}, nil
}

func RespToStr(resp *genai.GenerateContentResponse) string {
	return fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0])
}

func FilterComments(code string) string {
	commentStyles := []string{"//", "#"}
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

func (m *GenerationModel) CreateTests(code string) (string, error) {
	prompt := fmt.Sprintf(
		"Write unit tests in golang for the function with this header comment %s",
		code,
	)
	resp, err := m.model.GenerateContent(
		m.context,
		genai.Text(prompt),
	)
	if err != nil {
		return "", err
	}
	return RespToStr(resp), nil
}

func (m *GenerationModel) CreateComment(code string) (string, error) {
	prompt := fmt.Sprintf(
		"Write a comment for the following code, include details about parameters, return, and possible errors.\n Only write the function header comment, do not write any code.\n %s",
		code,
	)
	resp, err := m.model.GenerateContent(
		m.context,
		genai.Text(prompt),
	)
	if err != nil {
		return "", err
	}
	return FilterComments(RespToStr(resp)), nil
}
