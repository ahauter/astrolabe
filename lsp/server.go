package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

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
	log       commonlog.Logger
	version   string = "0.0.1"
	handler   protocol.Handler
	workspace Workspace
)
var model *GenerativeModelAPI

func MakeCommandHandler() (protocol.WorkspaceExecuteCommandFunc, error) {
	if model == nil {
		return func(context *glsp.Context, params *protocol.ExecuteCommandParams) (any, error) {
			return nil, nil
		}, errors.New("Cannot make command handler without model instantiated")
	}
	return func(context *glsp.Context, params *protocol.ExecuteCommandParams) (any, error) {
		log.Debug("Received Command")
		command := params.Command
		log.Debug(command)
		log.Debugf("%s", params.Arguments)
		switch command {
		case "create_comment":
			file_type := params.Arguments[0].(string)
			if file_type == "" {
				return "", errors.New("Invalid file type")
			}
			code := params.Arguments[len(params.Arguments)-1].(string)
			comment, err := model.CreateComment(file_type, code, commonlog.NewScopeLogger(log, "createComment"))
			return comment, err
		case "create_tests":
			comment := params.Arguments[0].(string)
			file_name := params.Arguments[1].(string)
			file_type := params.Arguments[2].(string)
			if file_type == "" {
				return "", errors.New("Invalid file type")
			}
			ext_index := strings.LastIndex(file_name, ".")
			if ext_index < 0 {
				return "", errors.New("Invalid file name")
			}
			file_name = file_name[:ext_index] + "_test" + file_name[ext_index:]
			tests, err := model.CreateTests(comment, file_type)
			tests = tests + "\n" + "__astro_test_file_path__=" + file_name + "\n"
			return tests, err
		case "run_diagnostics":
			return "", nil
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
	}, nil
}

func main() {
	// This increases logging verbosity (optional)
	path := "/home/austin/astrologs/lsp.astro.log"
	commonlog.Configure(1, &path)
	log = commonlog.GetLogger("lsp.main")
	log.SetMaxLevel(commonlog.Debug)
	model = &GenerativeModelAPI{
		endpoint: "http://127.0.0.1:8000/",
		model:    "loraModel",
	}
	CommandHandler, err := MakeCommandHandler()
	if err != nil {
		log.Criticalf("Error possibly from not instantiating your model, possibly not..")
	}
	handler = protocol.Handler{
		Initialize:              initialize,
		Initialized:             initialized,
		Shutdown:                shutdown,
		SetTrace:                setTrace,
		WorkspaceExecuteCommand: CommandHandler,
		TextDocumentCompletion:  textCompletion,
		TextDocumentDidChange:   workspace.HandleChange,
		TextDocumentDidSave:     fileSaveHandler,
	}
	server := server.NewServer(&handler, lsName, false)

	log.Info("Starting lsp server")
	server.RunStdio()
}

func textCompletion(
	context *glsp.Context,
	params *protocol.CompletionParams,
) (any, error) {
	log.Debug(*&params.TextDocument.URI)
	log.Debug("Completions")
	log.Debugf("Line: %s", params.Position.Line)
	log.Debugf("Col:  %s", params.Position.Character)
	startTime := time.Now()
	document, err := workspace.GetDocument(*&params.TextDocument.URI)
	result := []protocol.CompletionItem{}
	if err != nil {
		log.Debug("Error finding document for completion, exiting early")
		return result, nil
	}
	if model == nil {
		log.Debug("Model not instantiated, exiting early")
		return result, nil
	}
	txt, err := model.CodeCompletion(document.Lines(), int(params.Position.Line), int(params.Position.Character), log)
	log.Debugf("Completion time: {%s}", time.Since(startTime))

	if err != nil {
		log.Warning("No completion generated")
		log.Debug(err.Error())

		return result, nil
	}
	log.Debugf("Text generated: %s", txt)
	kind := protocol.CompletionItemKindText
	result = append(result, protocol.CompletionItem{
		TextEdit: protocol.TextEdit{
			NewText: txt,
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      params.Position.Line,
					Character: params.Position.Character,
				},
				End: protocol.Position{
					Line:      params.Position.Line,
					Character: params.Position.Character,
				},
			},
		},
		Kind: &kind,
	})
	log.Debugf("%i", len(result))
	list := protocol.CompletionList{
		Items:        result,
		IsIncomplete: true,
	}
	return list, nil
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()
	w, err := MakeWorkspace(params.WorkspaceFolders)
	if err != nil || w == nil {
		return "", err
	}
	workspace = *w
	log.Infof("Workspace size %i", workspace.Size())
	capabilities.CompletionProvider = &protocol.CompletionOptions{
		ResolveProvider: &protocol.True, // if you support resolve
	}
	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version,
		},
	}, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	log.Info("Initialized Server :)")
	return nil
}

func shutdown(context *glsp.Context) error {
	log.Info("Shutting down server")
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func fileSaveHandler(
	context *glsp.Context,
	params *protocol.DidSaveTextDocumentParams,
) error {
	doc, err := workspace.GetDocument(params.TextDocument.URI)
	if err != nil {
		return err
	}
	doc.Read()
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}
