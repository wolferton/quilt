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
	//"bytes"
	//"fmt"
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
	tokenisedTemplates map[string][]*QueryTemplateToken
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

func (qm *QueryManager) parseQueryFiles(files []string) map[string][]*QueryTemplateToken {

	fl := qm.FrameworkLogger
	tokenisedTemplates := map[string][]*QueryTemplateToken{}

	for _, filePath := range files {

		fl.LogDebug("Parsing query file " + filePath)

		file, err := os.Open(filePath)

		if err != nil {
			fl.LogError("Unable to open " + filePath + " for parsing: " + err.Error())
			continue
		}

		defer file.Close()

		scanner := bufio.NewScanner(file)
		idPrefix := qm.QueryIdPrefix
		trimId := qm.TrimIdWhiteSpace

		var queryTokens = []*QueryTemplateToken{}
		var currentToken *QueryTemplateToken = nil
		var id string

		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, idPrefix) {
				newId := strings.TrimPrefix(line, idPrefix)

				if trimId {
					newId = strings.TrimSpace(newId)
				}

				if id == "" {
					id = newId
				}

				if currentToken != nil {
					queryTokens = append(queryTokens, currentToken)
					tokenisedTemplates[id] = queryTokens
					queryTokens = []*QueryTemplateToken{}
					id = newId
				}

				currentToken = NewQueryTemplateToken(Fragment)
				queryTokens = append(queryTokens, currentToken)

			} else if !qm.blankLine(line) {

				re := regexp.MustCompile("\\$\\{([^\\}])\\}")

				matches := re.FindAllStringSubmatch(line, -1)

				if matches == nil {
					currentToken.AddContent(line + qm.NewLine)
				} else {
					fragments := re.Split(line, -1)

					firstMatch := re.FindStringIndex(line)

					startsWithVar := (firstMatch[0] == 0)
					varCount := len(matches)
					fragmentCount := len(fragments)

					for i := 0; i < varCount && i < fragmentCount; i++ {

						var varToken *QueryTemplateToken = nil
						var fragToken *QueryTemplateToken = nil

						if i < varCount {
							varLabel := matches[i][1]

							index, err := strconv.Atoi(varLabel)

							if err == nil {
								varToken = NewQueryTemplateToken(VarIndex)
								varToken.Index = index
							} else {
								varToken = NewQueryTemplateToken(VarName)
								varToken.Content = varLabel
							}
						}

						if i < fragmentCount {
							fragToken = NewQueryTemplateToken(Fragment)
							fragToken.AddContent(fragments[i])
						}

						if startsWithVar {
							queryTokens = qm.AddTokens(varToken, fragToken, queryTokens)

						} else {
							queryTokens = qm.AddTokens(fragToken, varToken, queryTokens)
						}

					}

					currentToken = queryTokens[len(queryTokens)-1]

					if currentToken.Type != Fragment {
						currentToken = NewQueryTemplateToken(Fragment)
						queryTokens = append(queryTokens, currentToken)
					}

					currentToken.AddContent(qm.NewLine)
				}

			}
		}

		if currentToken != nil {
			//fl.LogTrace("Storing query with id " + id)
			tokenisedTemplates[id] = queryTokens
		}
	}

	return tokenisedTemplates
}

func (qm *QueryManager) AddTokens(first *QueryTemplateToken, second *QueryTemplateToken, tokens []*QueryTemplateToken) []*QueryTemplateToken {

	if first != nil {
		tokens = append(tokens, first)
	}

	if second != nil {
		tokens = append(tokens, second)
	}

	return tokens
}

func (qm *QueryManager) blankLine(line string) bool {
	return len(strings.TrimSpace(line)) == 0
}

type QueryTokenType int

const (
	Fragment = iota
	VarName
	VarIndex
	Empty
)

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

func (qtt *QueryTemplateToken) AddContent(content string) {

	qtt.Content += content
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
