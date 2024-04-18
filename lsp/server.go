package main

import (
	"log"
	"os"

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

func CommandHandler(context *glsp.Context, params *protocol.ExecuteCommandParams) (any, error) {
	log.Println("Received Command")
	log.Println(params.Command)
	log.Println(params.Arguments)
	return "", nil
}

func main() {
	// This increases logging verbosity (optional)
	commonlog.Configure(1, nil)

	log_file, err := os.OpenFile("./dbg/logs.txt", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.SetOutput(log_file)
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
