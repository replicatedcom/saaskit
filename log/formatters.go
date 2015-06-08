package log

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
)

type ConsoleFormatter struct{}

func (cf *ConsoleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	result := &bytes.Buffer{}

	formattedTime := entry.Time.Format("2006/01/02 15:04:05")
	formattedLevel := strings.ToUpper(entry.Level.String())

	fmt.Fprintf(result, "%s %s %s %s", formattedLevel, formattedTime, entry.Data["replkit.file_loc"], entry.Message)

	for k, v := range entry.Data {
		if strings.HasPrefix(k, "replkit.") {
			continue
		}

		fmt.Fprintf(result, "\n\t%s: %v", k, v)
	}

	result.WriteByte('\n')
	return result.Bytes(), nil
}
