package main

import (
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

const lsName = "my language"

var (
	version string = "0.0.1"
	handler protocol.Handler
)

func MakeCommandHandler(model *gemini.GenerationModel) protocol.WorkspaceExecuteCommandFunc {
	return func(context *glsp.Context, params *protocol.ExecuteCommandParams) (any, error) {
		log.Println("Received Command")
		log.Println(params.Command)
		log.Println(params.Arguments)
		code := params.Arguments[len(params.Arguments)-1].(string)
		comment, err := model.CreateComment(code)
		return comment, err
	}
}

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
