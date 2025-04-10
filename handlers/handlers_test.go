package handlers

import (
    "testing"
    "net/http"
    "net/http/httptest"
)

func TestGetBalance_Success(t *testing.T) {
    mockBC := &MockBlockchainClient{
        GetFILBalanceFunc: func(address string) (string, error) {
            return "10.0", nil
        },
        GetIFILBalanceFunc: func(address string) (string, error) {
            return "5.0", nil
        },
    }

    req, _ := http.NewRequest("GET", "/balance/0x123", nil)
    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(handlers.GetBalance(mockBC))
    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    expected := `{"fil":"10.0","ifil":"5.0"}`
    if rr.Body.String() != expected {
        t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
    }
}
