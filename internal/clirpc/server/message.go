package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/wso2/integration-platform-tools/internal/auth"
	"github.com/wso2/integration-platform-tools/internal/clirpc/logger"
)

type ErrorCode int

type Result struct {
	Value interface{}
}

func CreateResult(value interface{}) Result {
	return Result{
		Value: value,
	}
}

type HandlerFunc func(*json.RawMessage) (Result, error)

var RequestSeperator = []byte{'\r', '\n', '\r', '\n'}

// RPC constants
const (
	ContentLengthHeader = "Content-Length"
	RPCVersion          = "2.0"
)

// type HanlderFunc func(param interface{}) (interface{}, error)

type RequestBase struct {
	RPC    string `json:"jsonrpc"`
	ID     int    `json:"id"`
	Method string `json:"method"`
}

type Request struct {
	RequestBase
	Params *json.RawMessage `json:"params"`
}

type ResponseBase struct {
	RPC string `json:"jsonrpc"`
	ID  *int   `json:"id"`
}

type Response struct {
	ResponseBase
	Result *interface{} `json:"result,omitempty"`
	Error  *Error       `json:"error,omitempty"`
}

type Notification struct {
	RPC    string `json:"rpc"`
	Method string `json:"method"`
}

type Error struct {
	Code ErrorCode   `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data,omitempty"`
}

type BaseRequest struct {
	Method string `json:"method"`
}

func NewResponse(id int, result interface{}) Response {
	return Response{
		ResponseBase: ResponseBase{
			RPC: RPCVersion,
			ID:  &id,
		},
		Result: &result,
	}
}

func EncodeMessage(msg any) string {
	content, err := json.Marshal(msg)

	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s: %d%s%s", ContentLengthHeader, len(content), string(RequestSeperator), content)
}

func DecodeMessage(data []byte) (req *Request, err error) {
	header, content, found := bytes.Cut(data, RequestSeperator)

	if !found {
		return req, fmt.Errorf("did not find the request seperator")
	}

	contentLenthBytes := header[len(fmt.Sprintf("%s: ", ContentLengthHeader)):]
	contentLength, err := strconv.Atoi(string(contentLenthBytes))

	if err != nil {
		return req, fmt.Errorf("invalid content length")
	}

	if contentLength != len(content) {
		return req, nil
	}

	if err := json.Unmarshal(content, &req); err != nil {
		return req, fmt.Errorf("invalid json")
	}

	if req.Method == "" {
		return req, fmt.Errorf("invalid method")
	}

	return req, nil
}

func SplitFunc(data []byte, _ bool) (advance int, token []byte, err error) {
	header, content, found := bytes.Cut(data, RequestSeperator)

	if !found {
		return 0, nil, nil
	}

	contentLenthBytes := header[len(fmt.Sprintf("%s: ", ContentLengthHeader)):]
	contentLength, err := strconv.Atoi(string(contentLenthBytes))

	if err != nil {
		return 0, nil, err
	}

	if len(content) < contentLength {
		return 0, nil, nil
	}

	totalLength := len(header) + len(RequestSeperator) + contentLength
	return totalLength, data[:totalLength], nil
}

type InitializeParams struct {
	ClientName    string `json:"clientName"`
	Version       string `json:"version"`
	CloudStsToken string `json:"cloudStsToken"`
}

type InitializeResult struct {
	ProcessID    int         `json:"processId"`
	Version      string      `json:"version"`
	Capabilities interface{} `json:"capabilities"`
}

func handleInitRequest(rm *json.RawMessage) (result Result, err error) {
	params := InitializeParams{}

	err = json.Unmarshal(*rm, &params)

	if err != nil {
		return Result{}, err
	}

	if params.CloudStsToken != "" {
		err := auth.SetInitToken(params.CloudStsToken, "0")
		if err != nil {
			return Result{}, err
		}
	}

	logger.Info("Received initialize request with client name: " + params.ClientName)

	result = CreateResult(InitializeResult{
		ProcessID:    1234,
		Version:      version,
		Capabilities: capabilities,
	})

	return result, nil
}
