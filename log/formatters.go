package log

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/replicatedcom/saaskit/param"
	"github.com/sirupsen/logrus"
)

type ConsoleFormatter struct{}

func (cf *ConsoleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	result := &bytes.Buffer{}

	formattedTime := entry.Time.Format("2006/01/02 15:04:05")
	formattedLevel := strings.ToUpper(entry.Level.String())

	if entry.Caller == nil {
		fmt.Fprintf(result, "%s %s %s", formattedLevel, formattedTime, entry.Message)
	} else {
		caller := fmt.Sprintf("%s:%d", shortPath(entry.Caller.File), entry.Caller.Line)
		fmt.Fprintf(result, "%s %s %s %s", formattedLevel, formattedTime, caller, entry.Message)
	}

	for k, v := range entry.Data {
		if strings.HasPrefix(k, "saaskit.") {
			continue
		}

		fmt.Fprintf(result, "\n\t%s: %v", k, v)
	}

	result.WriteByte('\n')
	return result.Bytes(), nil
}

func shortPath(pathIn string) string {
	projectName := param.Lookup("PROJECT_NAME", "", false)
	if projectName == "" || !strings.Contains(pathIn, projectName) {
		return pathIn
	}

	toks := strings.Split(pathIn, string(filepath.Separator))
	var resultToks []string
	for i := len(toks) - 1; i >= 0; i-- {
		t := toks[i]
		if t == projectName && i < len(toks)-1 {
			resultToks = toks[i+1:]
			break
		}
	}

	if resultToks == nil {
		return pathIn
	}

	// Truncate absurdly long paths even if they don't match our project name (e.g. godeps).
	if len(resultToks) > 4 {
		resultToks = resultToks[len(resultToks)-3:]
	}

	return strings.Join(resultToks, string(filepath.Separator))
}
