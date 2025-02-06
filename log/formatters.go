package log

import (
	"bytes"
	"encoding/json"
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

type JSONFormatter struct{}

var DefaultJSONFieldKeys = []string{"level", "timestamp", "caller", "message"}

func (jf *JSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+4)
	for k, v := range entry.Data {
		if strings.HasPrefix(k, "saaskit.") {
			continue
		}

		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	// Configure default fields.
	prefixDefaultFieldClashes(data)
	data["timestamp"] = entry.Time.Format("2006/01/02 15:04:05")
	data["level"] = entry.Level.String()
	data["message"] = entry.Message
	if entry.Caller != nil {
		data["caller"] = fmt.Sprintf("%s:%d", shortPath(entry.Caller.File), entry.Caller.Line)
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(true)
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON: %w", err)
	}

	return b.Bytes(), nil
}

// prefixDefaultFieldClashes adds a prefix to the keys in data that clash
// with the keys in DefaultJSONFieldKeys to prevent them from being overwritten.
func prefixDefaultFieldClashes(data logrus.Fields) {
	for _, fieldKey := range DefaultJSONFieldKeys {
		if _, ok := data[fieldKey]; ok {
			data["fields."+fieldKey] = data[fieldKey]
			// Delete the original non-prefixed key.
			delete(data, fieldKey)
		}
	}
}
