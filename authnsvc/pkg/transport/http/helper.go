package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	stdhttp "net/http"
	"strconv"
	"strings"

	stdjwt "github.com/dgrijalva/jwt-go"
	kitjwt "github.com/go-kit/kit/auth/jwt"
	kitep "github.com/go-kit/kit/endpoint"
	"gorm.io/gorm"

	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/dto"
	ce "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/error"
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
	// check error type here
	if _, ok := err.(*stdjwt.ValidationError); ok {
		return stdhttp.StatusBadRequest
	}

	switch err {
	case io.ErrUnexpectedEOF, io.EOF:
		return stdhttp.StatusBadRequest
	case ce.ErrWrongCred, ce.ErrTokenExpired, kitjwt.ErrTokenContextMissing, kitjwt.ErrTokenExpired:
		return stdhttp.StatusUnauthorized
	case ce.ErrInsufficientPerm:
		return stdhttp.StatusForbidden
	case gorm.ErrRecordNotFound:
		return stdhttp.StatusNotFound
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

func processBasicQP(r *stdhttp.Request) *dto.BasicQueryParam {
	l := dto.NewBasicQueryParam()
	if pageNo := r.FormValue("page"); pageNo != "" {
		if v, err := strconv.Atoi(pageNo); err == nil {
			l.Paginator.Page = v
		}
	}
	if pageSize := r.FormValue("page_size"); pageSize != "" {
		if v, err := strconv.Atoi(pageSize); err == nil {
			l.Paginator.PageSize = v
		}
	}
	orderBy := r.FormValue("orderby")
	if orderBy == "" {
		l.Filter.OrederBy = append(l.Filter.OrederBy, "updated_at desc")
	} else {
		fields := strings.Split(orderBy, ",")
		for _, f := range fields {
			fo := strings.Split(f, "__")
			if len(fo) > 2 {
				continue
			}
			l.Filter.OrederBy = append(l.Filter.OrederBy, strings.Join(fo, " "))
		}
	}
	return l
}
