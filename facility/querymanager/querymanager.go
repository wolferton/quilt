/*
Package querymanager provides and supports the QueryManager facility. The QueryManager provides a mechanism for
loading query templates from plain text files and allowing code to combine those templates with parameters to create a
query ready for execution.

The QueryManager is generic and is suitable for managing query templates for any data source that is interacted with via
text queries.
*/
package querymanager

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type QueryManager struct {
	TemplateLocation   string
	VarMatchRegEx      string
	FrameworkLogger    logger.Logger
	QueryIdPrefix      string
	TrimIdWhiteSpace   bool
	WrapStrings        bool
	StringWrapWith     string
	NewLine            string
	tokenisedTemplates map[string]*QueryTemplate
}

func (qm *QueryManager) StartComponent() {
	fl := qm.FrameworkLogger
	fl.LogDebug("Starting QueryManager")
	fl.LogDebug(qm.TemplateLocation)

	queryFiles, err := config.FileListFromPath(qm.TemplateLocation)

	if err == nil {

		qm.tokenisedTemplates = qm.parseQueryFiles(queryFiles)
		fl.LogDebug("Started QueryManager")
	} else {
		fl.LogFatal("Unable to start QueryManager due to problem loading query files: " + err.Error())
	}

}

func (qm *QueryManager) parseQueryFiles(files []string) map[string]*QueryTemplate {
	fl := qm.FrameworkLogger
	tokenisedTemplates := map[string]*QueryTemplate{}
	re := regexp.MustCompile(qm.VarMatchRegEx)

	for _, filePath := range files {

		fl.LogDebug("Parsing query file " + filePath)

		file, err := os.Open(filePath)

		if err != nil {
			fl.LogError("Unable to open " + filePath + " for parsing: " + err.Error())
			continue
		}

		defer file.Close()

		scanner := bufio.NewScanner(file)
		qm.scanAndParse(scanner, tokenisedTemplates, re)
	}

	return tokenisedTemplates
}

func (qm *QueryManager) scanAndParse(scanner *bufio.Scanner, tokenisedTemplates map[string]*QueryTemplate, re *regexp.Regexp) {

	var currentTemplate *QueryTemplate = nil
	var fragmentBuffer bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()

		idLine, id := qm.isIdLine(line)

		if idLine {

			if currentTemplate != nil {
				currentTemplate.Finalise()
			}

			currentTemplate = NewQueryTemplate(id, &fragmentBuffer)
			tokenisedTemplates[id] = currentTemplate
			continue
		}

		if qm.isBlankLine(line) {
			continue
		}

		varTokens := re.FindAllStringSubmatch(line, -1)

		if varTokens == nil {
			currentTemplate.AddFragmentContent(line)
		} else {

			fragments := re.Split(line, -1)
			firstMatch := re.FindStringIndex(line)

			startsWithVar := (firstMatch[0] == 0)
			varCount := len(varTokens)
			fragmentCount := len(fragments)

			for i := 0; i < varCount && i < fragmentCount; i++ {

				varAvailable := i < varCount
				fragAvailable := i < fragmentCount

				if varAvailable && fragAvailable {

					varToken := varTokens[i][1]
					fragment := fragments[i]

					if startsWithVar {
						qm.addVar(varToken, currentTemplate)
						currentTemplate.AddFragmentContent(fragment)
					} else {
						currentTemplate.AddFragmentContent(fragment)
						qm.addVar(varToken, currentTemplate)

					}

				} else if varAvailable {
					qm.addVar(varTokens[i][1], currentTemplate)

				} else if fragAvailable {
					currentTemplate.AddFragmentContent(fragments[i])
				}

			}
		}

		currentTemplate.EndLine()

	}

	if currentTemplate != nil {
		currentTemplate.Finalise()
	}

}

func (qm *QueryManager) addVar(token string, currentTemplate *QueryTemplate) {

	index, err := strconv.Atoi(token)

	if err == nil {
		currentTemplate.AddIndexedVar(index)
	} else {
		currentTemplate.AddLabelledVar(token)
	}
}

func (qm *QueryManager) isIdLine(line string) (bool, string) {
	idPrefix := qm.QueryIdPrefix

	if strings.HasPrefix(line, idPrefix) {
		newId := strings.TrimPrefix(line, idPrefix)

		if qm.TrimIdWhiteSpace {
			newId = strings.TrimSpace(newId)
		}

		return true, newId

	} else {
		return false, ""
	}
}

func (qm *QueryManager) isBlankLine(line string) bool {
	return len(strings.TrimSpace(line)) == 0
}

type QueryTokenType int

const (
	Fragment = iota
	VarName
	VarIndex
)

type QueryTemplate struct {
	Tokens         []*QueryTemplateToken
	Id             string
	currentToken   *QueryTemplateToken
	fragmentBuffer *bytes.Buffer
}

func (qt *QueryTemplate) Finalise() {
	qt.closeFragmentToken()
	qt.fragmentBuffer = nil
}

func (qt *QueryTemplate) AddFragmentContent(fragment string) {

	t := qt.currentToken

	if t == nil || t.Type != Fragment {
		t = NewQueryTemplateToken(Fragment)
		qt.Tokens = append(qt.Tokens, t)
		qt.currentToken = t
	}

	qt.fragmentBuffer.WriteString(fragment)
}

func (qt *QueryTemplate) closeFragmentToken() {

	t := qt.currentToken
	if t.Type == Fragment {
		t.Content = qt.fragmentBuffer.String()
		qt.fragmentBuffer.Reset()
	}

}

func (qt *QueryTemplate) AddIndexedVar(index int) {

	qt.closeFragmentToken()
	t := qt.currentToken

	t = NewQueryTemplateToken(VarIndex)
	t.Index = index

	qt.Tokens = append(qt.Tokens, t)
	qt.currentToken = t
}

func (qt *QueryTemplate) AddLabelledVar(label string) {

	qt.closeFragmentToken()
	t := qt.currentToken

	t = NewQueryTemplateToken(VarName)
	t.Content = label

	qt.Tokens = append(qt.Tokens, t)
	qt.currentToken = t
}

func (qt *QueryTemplate) EndLine() {
	qt.AddFragmentContent("\n")
}

func NewQueryTemplate(id string, buffer *bytes.Buffer) *QueryTemplate {
	t := new(QueryTemplate)
	t.Id = id
	t.currentToken = nil
	t.fragmentBuffer = buffer

	return t
}

type QueryTemplateToken struct {
	Type    QueryTokenType
	Content string
	Index   int
}

func NewQueryTemplateToken(tokenType QueryTokenType) *QueryTemplateToken {
	token := new(QueryTemplateToken)
	token.Type = tokenType

	return token
}

func (qtt *QueryTemplateToken) GetContent() string {
	return qtt.Content
}

func (qtt *QueryTemplateToken) String() string {

	switch qtt.Type {

	case Fragment:
		return qtt.Content
	case VarName:
		return fmt.Sprintf("VN:%s", qtt.Content)
	case VarIndex:
		return fmt.Sprintf("VI:%d", qtt.Index)
	default:
		return ""

	}
}
