package liclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type Client struct {
	Secret     string
	HTTPClient *http.Client
}

type State int

type Withdrawal struct {
	ID    string
	State State
	LNURL string
}

const (
	ReadyState State = iota
	ScannedState
	CallbackState
)

var (
	Host = "https://api.lnurl.it"

	UUIDRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}$`)
)

func (state State) String() string {
	switch state {
	case ReadyState:
		return "ready"

	case ScannedState:
		return "scanned"

	case CallbackState:
		return "callback"

	default:
		return "unknown"
	}
}

func (client *Client) fetch(method string, URL string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(method, URL, body)

	if err != nil {
		return nil, err
	}

	request.Header.Set("x-api-secret", client.Secret)

	return client.HTTPClient.Do(request)
}

func (client *Client) CreateWithdrawal(amount int, description string, webhookURL string) (Withdrawal, error) {
	buffer := new(bytes.Buffer)

	err := json.NewEncoder(buffer).Encode(struct {
		Amount      int    `json:"amount"`
		Description string `json:"description"`
		WebhookURL  string `json:"webhookURL"`
	}{
		Amount:      amount,
		Description: description,
		WebhookURL:  webhookURL,
	})

	if err != nil {
		return Withdrawal{}, err
	}

	response, err := client.fetch("POST", Host+"/v1/withdrawal/create", buffer)

	if err != nil {
		return Withdrawal{}, err
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		return Withdrawal{}, fmt.Errorf("expected either status 200 or 201: status code %d", response.StatusCode)
	}

	var body Withdrawal

	err = json.NewDecoder(response.Body).Decode(&body)

	if err != nil {
		return Withdrawal{}, err
	}

	body.State = ReadyState

	return body, nil
}

func (client *Client) GetWithdrawal(ID string) (Withdrawal, error) {
	if ID == "" {
		return Withdrawal{}, errors.New("ID cannot be empty")
	}

	if ok := UUIDRegex.MatchString(ID); !ok {
		return Withdrawal{}, errors.New("ID is invalid: must be in format of xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
	}

	response, err := client.fetch("GET", Host+"/v1/withdrawal/get?ID="+ID, nil)

	if err != nil {
		return Withdrawal{}, err
	}

	if response.StatusCode != http.StatusOK {
		return Withdrawal{}, fmt.Errorf("expected status 200: status code %d", response.StatusCode)
	}

	var body Withdrawal

	err = json.NewDecoder(response.Body).Decode(&body)

	if err != nil {
		return Withdrawal{}, err
	}

	body.ID = ID

	return body, nil
}

func (client *Client) DeleteWithdrawal(ID string) error {
	if ID == "" {
		return errors.New("ID cannot be empty")
	}

	if ok := UUIDRegex.MatchString(ID); !ok {
		return errors.New("ID is invalid: must be in format of xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
	}

	buffer := new(bytes.Buffer)

	err := json.NewEncoder(buffer).Encode(struct {
		ID string
	}{
		ID: ID,
	})

	if err != nil {
		return err
	}

	response, err := client.fetch("POST", Host+"/v1/withdrawal/delete", buffer)

	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status 200: status code %d", response.StatusCode)
	}

	return nil
}

func New(secret string) (Client, error) {
	if secret == "" {
		return Client{}, errors.New("secret cannot be empty")
	}

	if ok := UUIDRegex.MatchString(secret); !ok {
		return Client{}, errors.New("secret is invalid: must be in format of xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
	}

	return Client{
		Secret:     secret,
		HTTPClient: http.DefaultClient,
	}, nil
}
