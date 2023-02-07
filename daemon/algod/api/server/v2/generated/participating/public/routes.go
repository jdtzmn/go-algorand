// Package public provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/algorand/oapi-codegen DO NOT EDIT.
package public

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	. "github.com/algorand/go-algorand/daemon/algod/api/server/v2/generated/model"
	"github.com/algorand/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get a list of unconfirmed transactions currently in the transaction pool by address.
	// (GET /v2/accounts/{address}/transactions/pending)
	GetPendingTransactionsByAddress(ctx echo.Context, address string, params GetPendingTransactionsByAddressParams) error
	// Broadcasts a raw transaction or transaction group to the network.
	// (POST /v2/transactions)
	RawTransaction(ctx echo.Context) error
	// Get a list of unconfirmed transactions currently in the transaction pool.
	// (GET /v2/transactions/pending)
	GetPendingTransactions(ctx echo.Context, params GetPendingTransactionsParams) error
	// Get a specific pending transaction.
	// (GET /v2/transactions/pending/{txid})
	PendingTransactionInformation(ctx echo.Context, txid string, params PendingTransactionInformationParams) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetPendingTransactionsByAddress converts echo context to params.
func (w *ServerInterfaceWrapper) GetPendingTransactionsByAddress(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "address" -------------
	var address string

	err = runtime.BindStyledParameterWithLocation("simple", false, "address", runtime.ParamLocationPath, ctx.Param("address"), &address)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter address: %s", err))
	}

	ctx.Set(Api_keyScopes, []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetPendingTransactionsByAddressParams
	// ------------- Optional query parameter "max" -------------

	err = runtime.BindQueryParameter("form", true, false, "max", ctx.QueryParams(), &params.Max)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter max: %s", err))
	}

	// ------------- Optional query parameter "format" -------------

	err = runtime.BindQueryParameter("form", true, false, "format", ctx.QueryParams(), &params.Format)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter format: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetPendingTransactionsByAddress(ctx, address, params)
	return err
}

// RawTransaction converts echo context to params.
func (w *ServerInterfaceWrapper) RawTransaction(ctx echo.Context) error {
	var err error

	ctx.Set(Api_keyScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.RawTransaction(ctx)
	return err
}

// GetPendingTransactions converts echo context to params.
func (w *ServerInterfaceWrapper) GetPendingTransactions(ctx echo.Context) error {
	var err error

	ctx.Set(Api_keyScopes, []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetPendingTransactionsParams
	// ------------- Optional query parameter "max" -------------

	err = runtime.BindQueryParameter("form", true, false, "max", ctx.QueryParams(), &params.Max)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter max: %s", err))
	}

	// ------------- Optional query parameter "format" -------------

	err = runtime.BindQueryParameter("form", true, false, "format", ctx.QueryParams(), &params.Format)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter format: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetPendingTransactions(ctx, params)
	return err
}

// PendingTransactionInformation converts echo context to params.
func (w *ServerInterfaceWrapper) PendingTransactionInformation(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "txid" -------------
	var txid string

	err = runtime.BindStyledParameterWithLocation("simple", false, "txid", runtime.ParamLocationPath, ctx.Param("txid"), &txid)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter txid: %s", err))
	}

	ctx.Set(Api_keyScopes, []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params PendingTransactionInformationParams
	// ------------- Optional query parameter "format" -------------

	err = runtime.BindQueryParameter("form", true, false, "format", ctx.QueryParams(), &params.Format)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter format: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PendingTransactionInformation(ctx, txid, params)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface, m ...echo.MiddlewareFunc) {
	RegisterHandlersWithBaseURL(router, si, "", m...)
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string, m ...echo.MiddlewareFunc) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/v2/accounts/:address/transactions/pending", wrapper.GetPendingTransactionsByAddress, m...)
	router.POST(baseURL+"/v2/transactions", wrapper.RawTransaction, m...)
	router.GET(baseURL+"/v2/transactions/pending", wrapper.GetPendingTransactions, m...)
	router.GET(baseURL+"/v2/transactions/pending/:txid", wrapper.PendingTransactionInformation, m...)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
