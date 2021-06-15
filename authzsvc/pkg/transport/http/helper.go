package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	stdhttp "net/http"
	"strings"

	kitep "github.com/go-kit/kit/endpoint"
)

func ErrorEncoder(_ context.Context, err error, w stdhttp.ResponseWriter) {
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(hideSensitiveContent(errorWrapper{Error: err.Error()}))
}

func ErrorDecoder(r *stdhttp.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

type errorWrapper struct {
	Error string `json:"error"`
}

func hideSensitiveContent(e errorWrapper) errorWrapper {
	if strings.Contains(e.Error, "failed to connect") {
		e.Error = "server encountered some issue, please contact the Admin"
	}
	return e
}

// This is used to set the http status, see an example here :
// https://github.com/go-kit/kit/blob/master/examples/addsvc/pkg/addtransport/http.go#L133
func err2code(err error) int {
	switch err {
	case io.ErrUnexpectedEOF, io.EOF:
		return stdhttp.StatusBadRequest
	}
	return stdhttp.StatusInternalServerError
}

// encodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func encodeHTTPGenericResponse(ctx context.Context, w stdhttp.ResponseWriter, response interface{}) (err error) {
	if f, ok := response.(kitep.Failer); ok && f.Failed() != nil {
		ErrorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
