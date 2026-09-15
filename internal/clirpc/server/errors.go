package server

import (
	"errors"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

var (
	InvalidRequestError = errors.New("invalid request")
	MethodNotFoundError = errors.New("method not found")
	InvalidParamsError  = errors.New("invalid params")
	InternalErrorError  = errors.New("internal error")
	ParseErrorError     = errors.New("parse error")
)

// RPC error codes
const (
	InvalidRequest ErrorCode = -32600
	MethodNotFound ErrorCode = -32601
	InvalidParams  ErrorCode = -32602
	InternalError  ErrorCode = -32603
	ParseError     ErrorCode = -32700
	// ErrorCode represents a custom error code
	UnauthorizedError  ErrorCode = -32000
	TokenNotFoundError ErrorCode = -32001
	InvalidTokenError  ErrorCode = -32002
	ForbiddenError     ErrorCode = -32003
	RefreshTokenError  ErrorCode = -32004
	ComponentNotFound  ErrorCode = -32005
	ProjectNotFound    ErrorCode = -32006
	MaxProjectCount    ErrorCode = -32007
	RepoAccessNeeded   ErrorCode = -32008
	EpYamlNotFound     ErrorCode = -32009
	UserNotFound       ErrorCode = -32010
	MaxComponentCount  ErrorCode = -32011
	InvalidSubPath     ErrorCode = -32012
	NoOrgsAvailable    ErrorCode = -32013
	NoAccountAvailable ErrorCode = -32014
)

func NewErrorResponse(id int, err error) Response {
	code := InternalError

	if errors.Is(err, InvalidRequestError) {
		code = InvalidRequest
	} else if errors.Is(err, MethodNotFoundError) {
		code = MethodNotFound
	} else if errors.Is(err, InvalidParamsError) {
		code = InvalidParams
	} else if errors.Is(err, InternalErrorError) {
		code = InternalError
	} else if errors.Is(err, ParseErrorError) {
		code = ParseError
	} else if errors.Is(err, api.ErrNotLoggedIn) {
		code = UnauthorizedError
	} else if errors.Is(err, api.ErrTokenNotValid) {
		code = InvalidTokenError
	} else if errors.Is(err, api.ErrNoTokenFoundForOrg) {
		code = TokenNotFoundError
	} else if errors.Is(err, api.ErrForbidden) {
		code = ForbiddenError
	} else if errors.Is(err, api.ErrRefreshToken) {
		code = RefreshTokenError
	} else if errors.Is(err, api.ErrFailedToResolveComp) {
		code = ComponentNotFound
	} else if errors.Is(err, api.ErrFailedToResolveProj) {
		code = ProjectNotFound
	} else if errors.Is(err, api.ErrMaxProjectCountReached) {
		code = MaxProjectCount
	} else if errors.Is(err, api.RepoAccessNeeded) {
		code = RepoAccessNeeded
	} else if errors.Is(err, api.ComponentYamlNotFound) {
		code = EpYamlNotFound
	} else if errors.Is(err, api.UserNotFound) {
		code = UserNotFound
	} else if errors.Is(err, api.ErrMaxComponentCountReached) {
		code = MaxComponentCount
	} else if errors.Is(err, api.InvalidSubPath) {
		code = InvalidSubPath
	} else if errors.Is(err, api.NoOrgsAvailable) {
		code = NoOrgsAvailable
	} else if errors.Is(err, api.ErrNoAccountFound) {
		code = NoAccountAvailable
	}

	return Response{
		ResponseBase: ResponseBase{
			RPC: RPCVersion,
			ID:  &id,
		},
		Error: &Error{
			Code: code,
			Msg:  err.Error(),
		},
	}
}
