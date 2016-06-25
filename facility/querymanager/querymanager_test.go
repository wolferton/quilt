package querymanager

import (
	"bytes"
	"github.com/wolferton/quilt/facility/logger"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestSingleSingleQueryNoVars(t *testing.T) {

	queryFiles := []string{os.Getenv("QUILT_HOME") + "/test/querymanager/single-query-no-vars"}
	qm := buildQueryManager()

	tt := qm.parseQueryFiles(queryFiles)

	members := len(tt)

	if members != 1 {
		t.Errorf("Expected one entry in tokens map, found %d", members)
	}

}

func TestSingleQueryIndexVars(t *testing.T) {

	queryFiles := []string{os.Getenv("QUILT_HOME") + "/test/querymanager/single-query-index-vars"}
	qm := buildQueryManager()

	tt := qm.parseQueryFiles(queryFiles)

	members := len(tt)

	if members != 1 {
		t.Errorf("Expected one entry in tokens map, found %d", members)
		t.FailNow()
	}

	tokenisedQuery := tt["SINGLE_QUERY_INDEX_VARS"]

	if tokenisedQuery == nil {
		t.Errorf("Expected a query named SINGLE_QUERY_INDEX_VARS")
	}

	stringQuery := ToString(tokenisedQuery.Tokens)

	refQuery := LoadRefFile("/test/querymanager/single-query-index-vars-ref")

	if stringQuery != refQuery {
		t.Errorf("Generated query and reference query do not match. \nGEN:%s\nREF:%s\n", VisibleWhitespace(stringQuery), VisibleWhitespace(refQuery))
	}

}

func TestMultiQueryNameVars(t *testing.T) {

	queryFiles := []string{os.Getenv("QUILT_HOME") + "/test/querymanager/multi-query-name-vars"}
	qm := buildQueryManager()

	tt := qm.parseQueryFiles(queryFiles)

	members := len(tt)

	if members != 2 {
		t.Errorf("Expected two entries in tokens map, found %d", members)
		t.FailNow()
	}

	tokenisedQueryTwo := tt["MULTI_QUERY_INDEX_VARS_TWO"]

	if tokenisedQueryTwo == nil {
		t.Errorf("Expected a query named MULTI_QUERY_INDEX_VARS_TWO")
	}

	varsInQuery := NamedVarCount(tokenisedQueryTwo.Tokens)

	if varsInQuery != 3 {
		t.Errorf("Expected three named variables to be found in MULTI_QUERY_INDEX_VARS_TWO, found %d", varsInQuery)
	}

	stringQuery := ToString(tokenisedQueryTwo.Tokens)

	refQuery := LoadRefFile("/test/querymanager/multi-query-name-vars-ref-2")

	if stringQuery != refQuery {
		t.Errorf("Generated query and reference query do not match. \nGEN:%s\nREF:%s\n", VisibleWhitespace(stringQuery), VisibleWhitespace(refQuery))
	}
}

func NamedVarCount(tokens []*QueryTemplateToken) int {

	count := 0

	for _, token := range tokens {

		if token.Type == VarName {

			count += 1
		}

	}

	return count
}

func VisibleWhitespace(query string) string {

	nonewline := strings.Replace(query, "\n", "\\n", -1)

	return strings.Replace(nonewline, "\t", "\\t", -1)

}

func LoadRefFile(path string) string {
	bytes, _ := ioutil.ReadFile(os.Getenv("QUILT_HOME") + path)

	return string(bytes)
}

func ToString(tokens []*QueryTemplateToken) string {

	var buffer bytes.Buffer

	for _, token := range tokens {
		buffer.WriteString(token.String())
	}

	return buffer.String()
}

func buildQueryManager() *QueryManager {

	qm := new(QueryManager)
	qm.QueryIdPrefix = "ID:"
	qm.StringWrapWith = "'"
	qm.TrimIdWhiteSpace = true
	qm.VarMatchRegEx = "\\$\\{([^\\}]*)\\}"
	qm.NewLine = "\n"
	qm.FrameworkLogger = logger.CreateAnonymousLogger("querymanager_test", 0)

	return qm

}
