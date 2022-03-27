package query

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

type FileExtractor struct {
	Fs afero.Fs
}

type ExplainableQuery struct {
	VariableDefinitions []string
	Query               string
}

func (e ExplainableQuery) ToExplainQuery() string {
	eq := ""
	if len(e.VariableDefinitions) > 0 {
		eq += strings.Join(e.VariableDefinitions, ";\n") + ";\n"
	}

	eq += "EXPLAIN " + e.Query + ";"
	return eq
}

var queryCommentRegex = regexp.MustCompile(`(?s)\/\*.*?\*\/|--.*?\n`)

func (f FileExtractor) ExtractQueriesFromFile(filepath string) ([]*ExplainableQuery, error) {
	contents, err := afero.ReadFile(f.Fs, filepath)
	if err != nil {
		return nil, errors.Wrap(err, "could not read file")
	}

	cleanedUpQueries := queryCommentRegex.ReplaceAllLiteralString(string(contents), "\n")

	return splitQueries(cleanedUpQueries), nil
}

func splitQueries(fileContent string) []*ExplainableQuery {
	var queries []*ExplainableQuery
	var sqlVariablesSeenSoFar []string

	for _, query := range strings.Split(fileContent, ";") {
		query = strings.TrimSpace(query)
		if len(query) == 0 {
			continue
		}

		queryLines := strings.Split(query, "\n")
		cleanQueryRows := make([]string, 0, len(queryLines))
		for _, line := range queryLines {
			emptyLine := strings.TrimSpace(line)
			if len(emptyLine) == 0 {
				continue
			}

			cleanQueryRows = append(cleanQueryRows, line)
		}

		cleanQuery := strings.TrimSpace(strings.Join(cleanQueryRows, "\n"))
		lowerCaseVersion := strings.ToLower(cleanQuery)
		if strings.HasPrefix(lowerCaseVersion, "set") || strings.HasPrefix(lowerCaseVersion, "declare") {
			sqlVariablesSeenSoFar = append(sqlVariablesSeenSoFar, cleanQuery)
			continue
		}

		queries = append(queries, &ExplainableQuery{
			VariableDefinitions: sqlVariablesSeenSoFar,
			Query:               strings.TrimSpace(cleanQuery),
		})
	}

	return queries
}
