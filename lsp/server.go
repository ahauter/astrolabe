package main

import (
	"errors"
	"fmt"
	"log"

	gemini "gemini/gemini"

	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"

	// Must include a backend implementation
	// See CommonLog for other options: https://github.com/tliron/commonlog
	_ "github.com/tliron/commonlog/simple"
)

const lsName = "Astrolabe"

var (
	version string = "0.0.1"
	handler protocol.Handler
)

func MakeCommandHandler(model *gemini.GenerationModel) protocol.WorkspaceExecuteCommandFunc {
	return func(context *glsp.Context, params *protocol.ExecuteCommandParams) (any, error) {
		log.Println("Received Command")
		command := params.Command
		log.Println(command)
		log.Println(params.Arguments)
		switch command {
		case "create_comment":
			code := params.Arguments[len(params.Arguments)-1].(string)
			comment, err := model.CreateComment(code)
			return comment, err
		case "create_tests":
			comment := params.Arguments[0].(string)
			file_name := params.Arguments[1].(string)
			tests, err := model.CreateTests(comment)
			tests = tests + "\n" + "__astro_test_file_path__=" + file_name + "\n"
			return tests, err
		case "run_diagnostics":
			comment := params.Arguments[len(params.Arguments)-1].(string)
			tests, err := model.CreateTests(comment)
			return tests, err
		case "clear_diagnostics":
			clear_diagnostics := protocol.PublishDiagnosticsParams{
				URI:         "file:///home/austin/Repositories/Personal/astrolabe/lsp/server.go",
				Diagnostics: []protocol.Diagnostic{},
			}
			context.Notify("textDocument/publishDiagnostics", clear_diagnostics)
			return "", nil
		default:
			return "", errors.New(fmt.Sprintf("Unrecognized command type %s", command))
		}
	}
}

// MakeCommandHandler creates a CommandHandler for the given model.
//
// Parameters:
// - model: The model to use for the command handler.
//
// Returns:
// - A CommandHandler for the given model.
//
// Possible errors:
// - An error if the model is nil.
func main() {
	// This increases logging verbosity (optional)
	commonlog.Configure(1, nil)

	model, err := gemini.NewGenerationModel()
	if err != nil {
		log.Fatalf(err.Error())
	}
	CommandHandler := MakeCommandHandler(model)
	log.Println("Starting lsp server")
	handler = protocol.Handler{
		Initialize:              initialize,
		Initialized:             initialized,
		Shutdown:                shutdown,
		SetTrace:                setTrace,
		WorkspaceExecuteCommand: CommandHandler,
	}
	server := server.NewServer(&handler, lsName, false)

	server.RunStdio()
	log.Println("Starting lsp server")
}

// initialize provides the initialization parameters as
// defined in the language server protocol spec: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#initialize
// initializing the server initializes a connection with the client and creates the connection
// with specified capabilities.
func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()
	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version,
		},
	}, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	log.Println("Initialized Server")
	return nil
}

func shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	log.Println("Shutting down server")
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}
