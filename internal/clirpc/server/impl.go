package server

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wso2/integration-platform-tools/internal/clirpc/logger"
)

var version string = "DEV"

var handlerMap = make(map[string]HandlerFunc)
var capabilities = make(map[string]interface{})

// RegisterHandler registers a handler function for a given method
func RegisterHandler(method string, handler HandlerFunc) {
	// Register the handler
	handlerMap[method] = handler
	// Register the capability
	index := strings.Index(method, "/")

	if index == -1 {
		capabilities[method] = true
	} else {
		if _, ok := capabilities[method[:index]]; !ok {
			capabilities[method[:index]] = make(map[string]bool)
		}

		capabilities[method[:index]].(map[string]bool)[method[index+1:]] = true
	}

}

func StartRPCServer(v string) {
	if version != "" {
		version = v
	}

	scanner := bufio.NewScanner(os.Stdin)
	logger.Info("RPC server started. Waiting for incoming requests...")
	scanner.Split(SplitFunc) // splits the received messages based on the content length
	writer := os.Stdout
	for scanner.Scan() {
		data := scanner.Bytes()
		req, err := DecodeMessage(data)

		if err != nil {
			writeResponse(writer, NewErrorResponse(req.ID, ParseErrorError))
		}

		// handle the message
		handleMessge(writer, req)
	}
}

func RunSingleCmd(method string, params string) {
	param := fmt.Sprintf(`"params":%s`, params)

	rpcMsg := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"%s",%s}`, 0, method, param)
	message := fmt.Sprintf("%s: %d%s%s", ContentLengthHeader, len(rpcMsg), RequestSeperator, rpcMsg)

	req, err := DecodeMessage([]byte(message))

	if err != nil {
		writeResponse(os.Stdout, NewErrorResponse(req.ID, ParseErrorError))
		return
	}

	handleMessge(os.Stdout, req)
}

func handleMessge(writer io.Writer, req *Request) {
	// handle the message
	logger.Info("Received message with method: " + req.Method)

	if handler, ok := handlerMap[req.Method]; ok {
		result, err := handler(req.Params)

		if err != nil {
			writeResponse(writer, NewErrorResponse(req.ID, err))
			return
		}

		writeResponse(writer, NewResponse(req.ID, result.Value))
	} else {
		logger.Error("Didn't found method")
		writeResponse(writer, NewErrorResponse(req.ID, MethodNotFoundError))
		return
	}
}

func writeResponse(writer io.Writer, msg any) error {
	reply := EncodeMessage(msg)

	_, err := writer.Write([]byte(reply))

	if err != nil {
		return err
	}

	return nil
}

func init() {
	// register init handler
	RegisterHandler("initialize", handleInitRequest)
}
