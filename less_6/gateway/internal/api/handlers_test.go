package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/you/monorepo/gateway/internal/api"
	"github.com/you/monorepo/ledger"
)

func newTestServer() (*httptest.Server, *ledger.Ledger) {
	l := ledger.NewLedger()
	mux := http.NewServeMux()
	h := &api.Handlers{Ledger: l}
	h.Register(mux)
	return httptest.NewServer(mux), l
}

func mustDo(t *testing.T, c *http.Client, req *http.Request) *http.Response {
	t.Helper()
	res, err := c.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return res
}

func readBody(t *testing.T, r io.Reader) string {
	t.Helper()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}

func TestBudgetsEndpoints(t *testing.T) {
	ts, l := newTestServer()
	defer ts.Close()
	t.Cleanup(func() { l.Reset() })

	t.Run("post_ok", func(t *testing.T) {
		body := `{"category":"еда","limit":5000}`
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/budgets", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := mustDo(t, ts.Client(), req)
		defer res.Body.Close()

		if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Fatalf("want json content-type, got %q", ct)
		}
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("want 201, got %d (%s)", res.StatusCode, readBody(t, res.Body))
		}
	})

	t.Run("get_list_contains_item", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/budgets", nil)
		res := mustDo(t, ts.Client(), req)
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("want 200, got %d (%s)", res.StatusCode, readBody(t, res.Body))
		}

		var arr []struct {
			Category string  `json:"category"`
			Limit    float64 `json:"limit"`
		}
		if err := json.NewDecoder(res.Body).Decode(&arr); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(arr) == 0 || arr[0].Category != "еда" || arr[0].Limit != 5000 {
			t.Fatalf("unexpected list: %+v", arr)
		}
	})

	t.Run("post_bad_json", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/budgets", strings.NewReader("{bad json}"))
		req.Header.Set("Content-Type", "application/json")
		res := mustDo(t, ts.Client(), req)
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", res.StatusCode)
		}
	})

	t.Run("post_bad_limit", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/budgets", strings.NewReader(`{"category":"x","limit":0}`))
		req.Header.Set("Content-Type", "application/json")
		res := mustDo(t, ts.Client(), req)
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", res.StatusCode)
		}
	})
}

func TestTransactionsChain(t *testing.T) {
	ts, l := newTestServer()
	defer ts.Close()
	t.Cleanup(func() { l.Reset() })

	// prepare budget
	{
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/budgets", bytes.NewBufferString(`{"category":"еда","limit":2000}`))
		req.Header.Set("Content-Type", "application/json")
		res := mustDo(t, ts.Client(), req)
		res.Body.Close()
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("budget create failed: %d", res.StatusCode)
		}
	}

	t.Run("ok", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/transactions",
			bytes.NewBufferString(`{"amount":1500,"category":"еда","description":"ланч","date":"2025-09-10"}`))
		req.Header.Set("Content-Type", "application/json")
		res := mustDo(t, ts.Client(), req)
		defer res.Body.Close()

		if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Fatalf("want json content-type, got %q", ct)
		}
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("want 201, got %d (%s)", res.StatusCode, readBody(t, res.Body))
		}
	})

	t.Run("list", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/transactions", nil)
		res := mustDo(t, ts.Client(), req)
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("want 200, got %d", res.StatusCode)
		}
		var arr []map[string]any
		if err := json.NewDecoder(res.Body).Decode(&arr); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(arr) != 1 {
			t.Fatalf("want 1 tx, got %d", len(arr))
		}
	})

	t.Run("exceeded", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/transactions",
			bytes.NewBufferString(`{"amount":1000,"category":"еда"}`)) // 1500 + 1000 > 2000
		req.Header.Set("Content-Type", "application/json")
		res := mustDo(t, ts.Client(), req)
		defer res.Body.Close()

		if res.StatusCode != http.StatusConflict {
			t.Fatalf("want 409, got %d (%s)", res.StatusCode, readBody(t, res.Body))
		}
	})

	t.Run("bad_json", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/transactions", strings.NewReader("{oops"))
		req.Header.Set("Content-Type", "application/json")
		res := mustDo(t, ts.Client(), req)
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", res.StatusCode)
		}
	})
}
