package generated

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type GenericRequest struct {
	Params map[string]interface{}
}

type Transport interface {
	Call(method string, req GenericRequest) (interface{}, error)
}

type httpTransport struct {
	BaseURL string
	Client  *http.Client
}

func (t *httpTransport) Call(method string, req GenericRequest) (interface{}, error) {
	body, _ := json.Marshal(req)
	resp, err := t.Client.Post(t.BaseURL + "/" + method, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s", resp.Status)
	}
	
	var result interface{}
	// Decode logic... for now just simple
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Structs ---


type Client struct {
	transport Transport
	
}


func (c *Client) GetSystemStatus(req GenericRequest) (interface{}, error) {
	return c.transport.Call("libreria-a.system.GetSystemStatus", req)
}


type Client struct {
	transport Transport
	
}


func (c *Client) InternationalTransfer(req GenericRequest) (interface{}, error) {
	return c.transport.Call("libreria-a.transfers.international.InternationalTransfer", req)
}


type Client struct {
	transport Transport
	
}


func (c *Client) GetUserBalance(req GenericRequest) (interface{}, error) {
	return c.transport.Call("libreria-a.transfers.national.GetUserBalance", req)
}

func (c *Client) Transfer(req GenericRequest) (interface{}, error) {
	return c.transport.Call("libreria-a.transfers.national.Transfer", req)
}


type Client struct {
	transport Transport
	
	International *Client
	
	National *Client
	
}



type Client struct {
	transport Transport
	
	System *Client
	
	Transfers *Client
	
}



type Client struct {
	transport Transport
	
	Libreriaa *Client
	
}




func NewClient(baseURL string) *Client {
	t := &httpTransport{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
	c := &Client{transport: t}
	
	// Manually Init Knowledge (PoC)
	// Ideally this is recursively generated
	c.LibreriaA = &LibreriaAClient{transport: t}
	c.LibreriaA.System = &LibreriaASystemClient{transport: t}
	c.LibreriaA.Transfers = &LibreriaATransfersClient{transport: t}
	c.LibreriaA.Transfers.National = &LibreriaATransfersNationalClient{transport: t}
	c.LibreriaA.Transfers.International = &LibreriaATransfersInternationalClient{transport: t}

	return c
}
