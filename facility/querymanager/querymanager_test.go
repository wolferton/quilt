package querymanager

import (
	"bytes"
	"fmt"
	"github.com/wolferton/quilt/facility/logger"
	"os"
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

func TestSingleSingleQueryIndexVars(t *testing.T) {

	queryFiles := []string{os.Getenv("QUILT_HOME") + "/test/querymanager/single-query-index-vars"}
	qm := buildQueryManager()

	tt := qm.parseQueryFiles(queryFiles)

	members := len(tt)

	if members != 1 {
		t.Errorf("Expected one entry in tokens map, found %d", members)
	}

	tokenisedQuery := tt["SINGLE_QUERY_INDEX_VARS"]

	if tokenisedQuery == nil {
		t.Errorf("Expected a query named SINGLE_QUERY_INDEX_VARS")
	}

	stringQuery := ToString(tokenisedQuery)
	fmt.Print(stringQuery)

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
	qm.VarMatchRegEx = "\\$\\{[^\\}]\\}"
	qm.NewLine = "\n"
	qm.FrameworkLogger = logger.CreateAnonymousLogger("querymanager_test", 0)

	return qm

}
