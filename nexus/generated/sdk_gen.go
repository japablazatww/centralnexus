package generated

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTP:    &http.Client{},
	}
}


func (c *Client) GetUserBalance(req GetUserBalanceRequest) (interface{}, error) {
	body, _ := json.Marshal(req)
	resp, err := c.HTTP.Post(c.BaseURL+"/liba/GetUserBalance", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server error: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result["result"], nil
}

func (c *Client) Transfer(req TransferRequest) (interface{}, error) {
	body, _ := json.Marshal(req)
	resp, err := c.HTTP.Post(c.BaseURL+"/liba/Transfer", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server error: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result["result"], nil
}

func (c *Client) GetSystemStatus(req GetSystemStatusRequest) (interface{}, error) {
	body, _ := json.Marshal(req)
	resp, err := c.HTTP.Post(c.BaseURL+"/liba/GetSystemStatus", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server error: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result["result"], nil
}

