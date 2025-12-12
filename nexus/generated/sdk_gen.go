package generated

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)



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


type LibreriaaSystemClient struct {
	transport Transport
	
}


func (c *LibreriaaSystemClient) GetSystemStatus(req GenericRequest) (interface{}, error) {
	return c.transport.Call("libreria-a.system.GetSystemStatus", req)
}


type LibreriaaTransfersNationalClient struct {
	transport Transport
	
}


func (c *LibreriaaTransfersNationalClient) GetUserBalance(req GenericRequest) (interface{}, error) {
	return c.transport.Call("libreria-a.transfers.national.GetUserBalance", req)
}

func (c *LibreriaaTransfersNationalClient) Transfer(req GenericRequest) (interface{}, error) {
	return c.transport.Call("libreria-a.transfers.national.Transfer", req)
}

func (c *LibreriaaTransfersNationalClient) ComplexTransfer(req GenericRequest) (interface{}, error) {
	return c.transport.Call("libreria-a.transfers.national.ComplexTransfer", req)
}


type LibreriaaTransfersInternationalClient struct {
	transport Transport
	
}


func (c *LibreriaaTransfersInternationalClient) InternationalTransfer(req GenericRequest) (interface{}, error) {
	return c.transport.Call("libreria-a.transfers.international.InternationalTransfer", req)
}


type LibreriaaTransfersClient struct {
	transport Transport
	
	National *LibreriaaTransfersNationalClient
	
	International *LibreriaaTransfersInternationalClient
	
}



type LibreriaaClient struct {
	transport Transport
	
	System *LibreriaaSystemClient
	
	Transfers *LibreriaaTransfersClient
	
}



type LibreriabLoansClient struct {
	transport Transport
	
}


func (c *LibreriabLoansClient) CalculateLoan(req GenericRequest) (interface{}, error) {
	return c.transport.Call("libreria-b.loans.CalculateLoan", req)
}


type LibreriabClient struct {
	transport Transport
	
	Loans *LibreriabLoansClient
	
}



type Client struct {
	transport Transport
	
	Libreriaa *LibreriaaClient
	
	Libreriab *LibreriabClient
	
}




func NewClient(baseURL string) *Client {
	t := &httpTransport{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
	c := &Client{transport: t}
	
	// Dynamic Init
	c.Libreriaa = &LibreriaaClient{transport: t}
	c.Libreriaa.System = &LibreriaaSystemClient{transport: t}
	c.Libreriaa.Transfers = &LibreriaaTransfersClient{transport: t}
	c.Libreriaa.Transfers.National = &LibreriaaTransfersNationalClient{transport: t}
	c.Libreriaa.Transfers.International = &LibreriaaTransfersInternationalClient{transport: t}
	c.Libreriab = &LibreriabClient{transport: t}
	c.Libreriab.Loans = &LibreriabLoansClient{transport: t}

	return c
}
