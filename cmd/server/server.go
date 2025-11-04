package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/yourname/yokai-finder-mcp/internal/handler"
)

// ===== JSON-RPC 2.0 型 =====

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type response struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      any        `json:"id"`
	Result  any        `json:"result,omitempty"`
	Error   *respError `json:"error,omitempty"`
}

type respError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP params

type initResult struct {
	Capabilities map[string]any `json:"capabilities"`
}

var h = handler.New()

func main() {
	rd := bufio.NewReader(os.Stdin)
	for {
		req, err := readFramed(rd)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("read err:", err)
			return
		}
		go handle(req)
	}
}

func handle(b []byte) {
	var req request
	if err := json.Unmarshal(b, &req); err != nil {
		writeResp(response{JSONRPC: "2.0", ID: nil, Error: &respError{Code: -32700, Message: "parse error"}})
		return
	}
	switch req.Method {
	case "initialize":
		writeResp(response{JSONRPC: "2.0", ID: req.ID, Result: initResult{Capabilities: map[string]any{"tools": map[string]any{"list": true, "call": true}}}})
	case "tools/list":
		writeResp(response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"tools": h.Tools()}})
	case "tools/call":
		// params: { name: string, arguments: object }
		var p struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			writeResp(response{JSONRPC: "2.0", ID: req.ID, Error: &respError{Code: -32602, Message: "invalid params"}})
			return
		}
		res, err := h.Call(nil, p.Name, p.Arguments)
		if err != nil {
			writeResp(response{JSONRPC: "2.0", ID: req.ID, Error: &respError{Code: -32000, Message: err.Error()}})
			return
		}
		writeResp(response{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"content": []any{map[string]any{"type": "json", "data": res}}}})
	default:
		writeResp(response{JSONRPC: "2.0", ID: req.ID, Error: &respError{Code: -32601, Message: "method not found"}})
	}
}

// ===== Framing: Content-Length ヘッダ =====

func readFramed(r *bufio.Reader) ([]byte, error) {
	// ヘッダ部
	var clen int
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break // 空行でヘッダ終わり
		}
		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			v := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			n, _ := strconv.Atoi(v)
			clen = n
		}
	}
	if clen <= 0 {
		return nil, fmt.Errorf("missing content-length")
	}
	buf := make([]byte, clen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func writeResp(res response) {
	b, _ := json.Marshal(res)
	var out bytes.Buffer
	fmt.Fprintf(&out, "Content-Length: %d\r\n\r\n", len(b))
	out.Write(b)
	os.Stdout.Write(out.Bytes())
}
