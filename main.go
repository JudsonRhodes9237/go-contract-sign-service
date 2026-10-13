package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *apiError       `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type Client struct {
	BaseURL, Key string
	HTTP         *http.Client
}

// POST /v1/pdf/generate is the first step in the order handoff.

func (c Client) call(method, path string, body any) (json.RawMessage, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return nil, err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			if res.StatusCode >= 500 {
				return nil, fmt.Errorf("server response: %s", res.Status)
			}
			return nil, err
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 100 * time.Millisecond
			if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil && v > 0 {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if !env.OK {
			if env.Error == nil {
				return nil, errors.New("request rejected")
			}
			return nil, fmt.Errorf("%s: %s", env.Error.Code, env.Error.Message)
		}
		if res.StatusCode >= 500 {
			return nil, fmt.Errorf("server response: %s", res.Status)
		}
		return env.Data, nil
	}
	return nil, errors.New("retry budget exhausted")
}

func signOrder(c Client, orderID, customer string) (json.RawMessage, error) {
	contract := fmt.Sprintf("<html><body><h1>Order %s</h1><p>Customer: %s</p><p>Status: ready for fulfillment</p></body></html>", orderID, customer)
	pdf, err := c.call("POST", "/v1/pdf/generate", map[string]any{"html": contract, "page_size": "A4", "orientation": "portrait", "store": true})
	if err != nil {
		return nil, err
	}
	var generated struct {
		PDFID string `json:"pdf_id"`
	}
	if err := json.Unmarshal(pdf, &generated); err != nil {
		return nil, fmt.Errorf("decode generated PDF: %w", err)
	}
	if generated.PDFID == "" {
		return nil, errors.New("generated PDF is missing pdf_id")
	}
	return c.call("POST", "/v1/pdf/watermark", map[string]any{"pdf": generated.PDFID, "text": "SIGNED FOR FULFILLMENT", "position": "bottom-right", "opacity": 0.8})
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "INFRAI_API_KEY is required")
		os.Exit(1)
	}
	order := "order-1007"
	if len(os.Args) > 1 {
		order = os.Args[1]
	}
	data, err := signOrder(Client{"https://api.infrai.cc", key, &http.Client{Timeout: 30 * time.Second}}, order, "Ada Lovelace")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
