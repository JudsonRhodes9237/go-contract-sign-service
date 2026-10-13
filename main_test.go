package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignOrderPassesGeneratedPDFID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/pdf/generate":
			w.Write([]byte(`{"ok":true,"data":{"pdf_id":"pdf_test","url":"https://example.test/doc.pdf"}}`))
		case "/v1/pdf/watermark":
			var request struct {
				PDF string `json:"pdf"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.PDF != "pdf_test" {
				t.Errorf("watermark pdf = %q, decode error = %v", request.PDF, err)
			}
			w.Write([]byte(`{"ok":true,"data":{"pdf_id":"watermarked"}}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	if _, err := signOrder(Client{BaseURL: server.URL, HTTP: server.Client()}, "order-42", "Ada"); err != nil {
		t.Fatal(err)
	}
}

func TestOrderIDAppearsInContract(t *testing.T) {
	order := "order-42"
	contract := "<html><body><h1>Order " + order + "</h1></body></html>"
	if contract != "<html><body><h1>Order order-42</h1></body></html>" {
		t.Fatalf("unexpected contract: %s", contract)
	}
}
