package gemini

import (
	"fmt"
	"log"
	"os"
	"strings"

	"context"

	"github.com/google/generative-ai-go/genai"
	dotEnv "github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	fmt.Println("Hello world")
	err := dotEnv.Load(".env")
	if err != nil {
		fmt.Println("Failed to load env file")
	}
	api_key := os.Getenv("API_KEY")
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(api_key))
	if err != nil {
		log.Fatalf("Could not create genia client")
	}
	defer client.Close()
	model := client.GenerativeModel("gemini-pro")
	if err != nil {
		log.Fatal(err)
	}
	comment, err := CreateComment(
		ctx, *model,
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

	fmt.Println(CreateTests(ctx, *model, comment))
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

func CreateTests(ctx context.Context, model genai.GenerativeModel, code string) (string, error) {
	prompt := fmt.Sprintf(
		"Write unit tests in golang for the function with this header comment",
		code,
	)
	resp, err := model.GenerateContent(
		ctx,
		genai.Text(prompt),
	)
	if err != nil {
		return "", err
	}
	return RespToStr(resp), nil
}

func CreateComment(ctx context.Context, model genai.GenerativeModel, code string) (string, error) {
	prompt := fmt.Sprintf(
		"Write a comment for the following code, include details about parameters, return, and possible errors.\n Only write the function header comment, do not write any code.\n %s",
		code,
	)
	resp, err := model.GenerateContent(
		ctx,
		genai.Text(prompt),
	)
	if err != nil {
		return "", err
	}
	return FilterComments(RespToStr(resp)), nil
}
