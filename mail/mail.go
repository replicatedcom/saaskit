package mail

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

func SendMail(recipient string, template string, context map[string]interface{}) error {
	if len(os.Getenv("MAIL_API")) == 0 {
		return errors.New("MAIL_API envvar must be set prior to using the mail api")
	}
	url := fmt.Sprintf("%s/v1/send", os.Getenv("MAIL_API"))

	type Request struct {
		Recipient string                 `json:"recipient"`
		Template  string                 `json:"template"`
		Context   map[string]interface{} `json:"context"`
	}
	request := Request{
		Recipient: recipient,
		Template:  template,
		Context:   context,
	}

	b, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		return nil
	}

	return fmt.Errorf("Unexpected response from mail-api: %d", resp.StatusCode)
}
