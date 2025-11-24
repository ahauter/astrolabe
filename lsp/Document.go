package main

import (
	"errors"
	"os"
	"strings"

	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const path = "/home/austin/astrologs/lsp.astro.log"

func readDocument(uri string) (string, error) {
	content, err := os.ReadFile(uri)
	if err != nil {
		return "", err
	}
	return string(content), err
}

type Workspace struct {
	documents []Document
}

func listDir(log commonlog.Logger, path string) ([]Document, error) {
	log.Debugf("listings the directory %s", path)
	entries, err := os.ReadDir(path)
	log.Debugf("%d", len(entries))
	var docs []Document
	if err != nil {
		return docs, nil
	}
	for _, entry := range entries {
		new_path := path + string(os.PathSeparator) + entry.Name()
		log.Debug(new_path)
		if entry.IsDir() {
			if entry.Name() == "venv" || entry.Name() == "node_modules" {
				return docs, nil
			}
			//TODO exclude common baddies
			more_docs, err := listDir(log, new_path)
			if err == nil {
				log.Warningf("Could not add folder %s", new_path)
			}

			for _, doc := range more_docs {
				docs = append(docs, doc)
			}
		} else {
			//select only .py, .go, .cs, etx
			log.Debugf("Found file %f", new_path)
			doc, err := MakeDocument(new_path)
			if err != nil {
				log.Errorf("Error reading document! %s", err)
			} else {
				docs = append(docs, doc)
			}
		}
	}
	return docs, nil
}

/*
MakeWorkspace creates a new workspace with the given workspace folders.

Parameters:
- folders: An array of workspace folders.

Returns:
- A pointer to the created workspace, nil on error.
- error: An error occurred during the execution, nil on success.

Possible Errors:
- Not a supported URI: If any of the folders' URIs do not start with "file://".
*/
func MakeWorkspace(folders []protocol.WorkspaceFolder) (*Workspace, error) {
	log.Debug("Making workspace")
	var documents []Document
	for _, f := range folders {
		path := f.URI
		fp_prefix := "file://"
		if !strings.HasPrefix(path, fp_prefix) {
			return nil, errors.New("Not a supported URI")
		}
		path = strings.TrimPrefix(path, fp_prefix)
		result, err := listDir(log, path)
		if err != nil {
			log.Warningf("Got errors reading workspace %s", f.Name)
		}
		for _, doc := range result {
			documents = append(documents, doc)
		}
	}

	return &Workspace{documents}, nil
}

func (w *Workspace) Size() int {
	if w == nil {
		return 0
	}
	return len(w.documents)
}

func (w *Workspace) GetDocument(uri string) (*Document, error) {
	uri = strings.TrimPrefix(uri, "file://")
	if w == nil {
		return nil, errors.New("Need a workspace for documents!")
	}
	for i := range w.documents {
		if w.documents[i].uri == uri {
			return &w.documents[i], nil
		}
	}
	return nil, errors.New("Document not found")
}

func (w *Workspace) HandleChange(
	context *glsp.Context,
	params *protocol.DidChangeTextDocumentParams,
) error {
	uri := strings.TrimPrefix(params.TextDocument.URI, "file://")
	if w == nil {
		return errors.New("Need a workspace for documents!")
	}
	for i := range w.documents {
		doc := w.documents[i]
		if doc.uri == uri {
			log.Debugf("Document found! %s", doc.uri)
			return doc.HandleChange(context, params)
		}
	}
	log.Debugf("Document not found! %s", params.TextDocument.URI)
	return nil
}

type Document struct {
	uri   string
	lines []string
}

func MakeDocument(uri string) (Document, error) {
	content, err := readDocument(uri)
	return Document{
		uri:   uri,
		lines: strings.Split(content, "\n"),
	}, err
}

func (d *Document) Read() {
	content, err := readDocument(d.uri)
	if err != nil {
		log.Warning("Error reading file!")
		log.Debug(d.uri)
	}
	if strings.Join(d.lines, "\n") != content {
		log.Debug("Current text does not match!")
	}
	d.lines = strings.Split(content, "\n")
}

func (d *Document) Line(row int) (string, error) {
	if len(d.lines) > row && row <= 0 {
		return d.lines[row], nil
	}
	return "", errors.New("Out of bounds error!!")
}

func (d *Document) Pos(row int, col int) (string, error) {
	line, err := d.Line(row)
	if err != nil {
		return "", err
	}
	strArr := []rune(line)
	if len(strArr) > col && col <= 0 {
		return string(strArr[col]), nil
	}
	return "", errors.New("Out of bounds error!!")
}

// EditRange edits a specified range within the document with the given text.
//
// Parameters:
//
//	startLine - The starting line number (0-indexed) of the range to edit.
//	  Must be a non-negative integer.
//	startCol - The starting column number (0-indexed) of the range to edit.
//	  Must be a non-negative integer.
//	endLine - The ending line number (0-indexed) of the range to edit.
//	  Must be a non-negative integer and greater than or equal to startLine.
//	endCol - The ending column number (0-indexed) of the range to edit.
//	  Must be a non-negative integer and greater than or equal to startCol.
//	text - The string text to insert into the document at the specified range.
//
// Returns:
//
//	An error if any of the provided parameters are out of bounds or if the
//	operation fails. Returns nil if the edit is successful.
//
// Errors:
//   - "Out of bounds!" - If startLine or startCol is negative, or if endLine
//     or endCol is out of the valid range for the document.
//   - Any other errors that may
func (d *Document) editRange(
	startLine int,
	startCol int,
	endLine int,
	endCol int,
	text string,
) error {
	log.Debugf("start l,c %i,%i ", startLine, startCol)
	log.Debugf("end l,c %i,%i ", endLine, endCol)
	log.Debugf("Text: %s", text)
	//split
	new_lines := []string{}
	first_nl := ""
	last_nl := ""
	if len(text) > 0 {
		new_lines = strings.Split(text, "\n")
		first_nl = new_lines[0]
		last_nl = new_lines[len(new_lines)-1]
	}
	if startLine < 0 {
		log.Error("Start line is negative!")
		return errors.New("Out of bounds! ")
	}

	if startCol < 0 {
		log.Error("Start Col is negative!")
		return errors.New("Out of bounds! ")
	}
	pre_lines := []string{}
	if startLine < len(d.lines) {
		pre_lines = d.lines[:startLine]
		start_line_text := d.lines[startLine]
		if startCol < len(start_line_text) {
			prefix := start_line_text[startCol:]
			if len(new_lines) == 0 {
				last_nl = prefix + first_nl
			} else {
				new_lines[0] = prefix + first_nl
			}
		}
	}

	log.Debug("Prefix computed")
	suf_lines := []string{}
	if endLine > 0 && endLine < len(d.lines) {
		end_line := d.lines[endLine]
		if endCol > 0 && endCol < len(end_line) {
			end_suf := end_line[endCol:]
			log.Debug("end ln suffix computed")
			if len(new_lines) == 0 {
				log.Debug("empty list")
				new_lines = append(new_lines, last_nl+end_suf)
			} else if len(new_lines) == 1 {
				log.Debug("len 1 list")
				new_lines[len(new_lines)-1] = new_lines[0] + end_suf
			} else {
				new_lines[len(new_lines)-1] = last_nl + end_suf
				log.Debug("len n list")
			}
		}
		log.Debug("Suffix computed")
	}

	log.Debug("Insertion computed")
	updated := append(pre_lines, append(new_lines, suf_lines...)...)
	log.Debug("Update computed")

	d.lines = updated
	return nil
}

func (d *Document) HandleChange(
	context *glsp.Context,
	params *protocol.DidChangeTextDocumentParams,
) error {
	log.Debug("Change command received")
	uri := strings.TrimPrefix(params.TextDocument.URI, "file://")
	if uri != d.uri {
		log.Error("Error URI mismatch")
		return errors.New("Update wrong document!")
	}
	for _, c := range params.ContentChanges {
		c := c.(protocol.TextDocumentContentChangeEvent)
		log.Debugf("Change command sent {}", c)
		start := c.Range.Start
		end := c.Range.End
		txt := c.Text
		d.editRange(
			int(start.Line),
			int(start.Character),
			int(end.Line),
			int(end.Character),
			txt,
		)
	}
	return nil
}
