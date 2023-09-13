package shell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Checkmarx/kics/pkg/model"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"mvdan.cc/sh/v3/syntax"
)

// Parser - parser for Proto files
type Parser struct {
}

// Resource separates the list of commands by file
type Resource struct {
	CommandList map[string][]Command `json:"command"`
}

// Command is the struct for each Buildah command
type Command struct {
	Cmd       string
	Original  string
	Value     string
	StartLine int `json:"_kics_line"`
	EndLine   int
}

// FromValue is the struct for each from
type FromValue struct {
	Value string
	Line  int
}

// Info has the relevant information to Buildah parser
type Info struct {
	IgnoreLines      []int
	From             map[string][]Command
	FromValues       []FromValue
	IgnoreBlockLines []int
}

// Parse - parses sehll to Json
func (p *Parser) Parse(_ string, fileContent []byte) ([]model.Document, []int, error) {
	var info Info
	info.From = map[string][]Command{}

	reader := bytes.NewReader(fileContent)
	f, err := syntax.NewParser(syntax.KeepComments(true)).Parse(reader, "")

	if err != nil {
		return nil, []int{}, err
	}

	ignoreLines := make([]int, 0)
	syntax.Walk(f, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.Stmt:
			info.getStmt(x)
		case *syntax.Comment:
			ignoreLines = append(ignoreLines, int(x.Hash.Line()))
		}
		return true
	})

	var documents []model.Document
	var resource Resource
	resource.CommandList = info.From
	doc := &model.Document{}
	j, err := json.Marshal(resource)
	if err != nil {
		return nil, []int{}, errors.Wrap(err, "failed to Marshal Shell")
	}

	err = json.Unmarshal(j, &doc)
	if err != nil {
		return nil, []int{}, errors.Wrap(err, "failed to Unmarshal Shell")
	}

	documents = append(documents, *doc)

	fmt.Println("---parse shell---")
	out, _ := json.Marshal(&documents)
	fmt.Println(string(out))
	fmt.Println(">>>parse shell<<<")
	return documents, ignoreLines, nil
}

// GetKind returns the kind of the parser
func (p *Parser) GetKind() model.FileKind {
	return model.KindSHELL
}

// SupportedExtensions returns sehll extensions
func (p *Parser) SupportedExtensions() []string {
	return []string{".sh"}
}

// SupportedTypes returns types supported by this parser, which are sehll
func (p *Parser) SupportedTypes() map[string]bool {
	return map[string]bool{"shell": true}
}

// GetCommentToken return the comment token of Docker - #
func (p *Parser) GetCommentToken() string {
	return "#"
}

// StringifyContent converts original content into string formatted version
func (p *Parser) StringifyContent(content []byte) (string, error) {
	return string(content), nil
}

// Resolve resolves proto files variables
func (p *Parser) Resolve(fileContent []byte, _ string) ([]byte, error) {
	return fileContent, nil
}

// GetResolvedFiles returns the list of files that are resolved
func (p *Parser) GetResolvedFiles() map[string]model.ResolvedFile {
	return make(map[string]model.ResolvedFile)
}

func (i *Info) getStmt(stmt *syntax.Stmt) {
	if cmd, ok := stmt.Cmd.(*syntax.CallExpr); ok {
		args := cmd.Args

		// get kics-scan ignore-block related to command + get command
		stCommand := i.getStmtInfo(stmt, args)

		fromValue := FromValue{
			Value: stCommand.Value,
			Line:  stCommand.StartLine,
		}
		i.FromValues = append(i.FromValues, fromValue)

		if stCommand.Cmd != "" {
			if len(i.FromValues) != 0 {
				v := i.FromValues[len(i.FromValues)-1].Value
				i.From[v] = append(i.From[v], stCommand)
			}
		}
	}
}

func (i *Info) getStmtInfo(stmt *syntax.Stmt, args []*syntax.Word) Command {
	var command Command
	minimumArgs := 2

	if len(args) > minimumArgs {
		cmd := strings.TrimSpace(getWordValue(args[1]))
		fullCmd := strings.TrimSpace(getFullCommand(args))
		value := strings.TrimPrefix(fullCmd, cmd)
		start := int(args[0].Pos().Line())
		end := int(args[len(args)-1].End().Line())

		command = Command{
			Cmd:       cmd,
			Original:  fullCmd,
			StartLine: start,
			EndLine:   end,
			Value:     strings.TrimSpace(value),
		}

		return command
	}
	return command
}

func getWordValue(wd *syntax.Word) string {
	printer := syntax.NewPrinter()
	var buf bytes.Buffer

	err := printer.Print(&buf, wd)

	if err != nil {
		log.Debug().Msgf("failed to get word value: %s", err)
	}

	value := buf.String()
	buf.Reset()

	return value
}

func getFullCommand(args []*syntax.Word) string {
	var buf bytes.Buffer
	printer := syntax.NewPrinter()

	call := &syntax.CallExpr{Args: args}

	err := printer.Print(&buf, call)

	if err != nil {
		log.Debug().Msgf("failed to get full command: %s", err)
	}

	command := buf.String()
	buf.Reset()

	command = strings.Replace(command, "\n", "", -1)
	command = strings.Replace(command, "\r", "", -1)
	command = strings.Replace(command, "\t", "", -1)
	command = strings.Replace(command, "\\", "", -1)

	return command
}
