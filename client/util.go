package client

import (
	"fmt"
	"net/http"
)

const (
	X_DENO_RAY = "x-deno-ray"
)

type APIResponse interface {
	StatusCode() int
}

// RespIsError returns true if the response status code is >= 400.
func RespIsError(resp APIResponse) bool {
	return resp.StatusCode() >= 400
}

// APIErrorDetail returns a string with the error details from the API response.
// TODO: We could make the parameter simpler by modifying a template used for
// API client generation to add `GetBody` and `GetHeaders` methods to every
// response structs.
// See https://github.com/deepmap/oapi-codegen/issues/240
func APIErrorDetail(resp *http.Response, body []byte) string {
	if resp == nil {
		return "failed to extract API error detail"
	}

	traceID := resp.Header.Get(X_DENO_RAY)
	if traceID == "" {
		traceID = "<unknown>"
	}

	return fmt.Sprintf("API request errored with status code %d.\nResponse body: %s\n\nPlease contact the support team with the ID: %s.", resp.StatusCode, body, traceID)
}
