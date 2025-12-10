package generated

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	BaseURL    string
	HTTP       *http.Client
	LibreriaA  *LibreriaAClient
}

func NewClient(baseURL string) *Client {
	c := &Client{
		BaseURL: baseURL,
		HTTP:    &http.Client{},
	}
	c.LibreriaA = &LibreriaAClient{client: c}
	return c
}

type LibreriaAClient struct {
	client *Client
}


func (c *LibreriaAClient) GetUserBalance(req GenericRequest) (interface{}, error) {
	body, _ := json.Marshal(req)
	resp, err := c.client.HTTP.Post(c.client.BaseURL+"/liba/GetUserBalance", "application/json", bytes.NewBuffer(body))
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

func (c *LibreriaAClient) Transfer(req GenericRequest) (interface{}, error) {
	body, _ := json.Marshal(req)
	resp, err := c.client.HTTP.Post(c.client.BaseURL+"/liba/Transfer", "application/json", bytes.NewBuffer(body))
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

func (c *LibreriaAClient) GetSystemStatus(req GenericRequest) (interface{}, error) {
	body, _ := json.Marshal(req)
	resp, err := c.client.HTTP.Post(c.client.BaseURL+"/liba/GetSystemStatus", "application/json", bytes.NewBuffer(body))
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

