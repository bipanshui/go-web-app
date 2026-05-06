// Test the main function

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMain(t *testing.T) {
	req, err := http.NewRequest("GET", "/home", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homePage)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Just verify the code not html content
	expected := "text/html; charset=utf-8"
	if contentType := rr.Header().Get("Content-Type"); contentType != expected {
		t.Errorf("handler returned unexpected content type: got %v want %v",
			contentType, expected)
	}
}

func TestRootRedirectsToHome(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusFound {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}

	if location := rr.Header().Get("Location"); location != "/home" {
		t.Fatalf("handler returned unexpected redirect target: got %v want %v", location, "/home")
	}
}

func TestStaticStylesheetServed(t *testing.T) {
	req, err := http.NewRequest("GET", "/static/site.css", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	newMux().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	if len(body) == 0 {
		t.Fatal("expected stylesheet body to be non-empty")
	}
}
