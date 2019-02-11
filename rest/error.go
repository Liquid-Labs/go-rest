package rest

import (
  "errors"
  "fmt"
  "net/http"
  "runtime"
)

type RestError interface {
  Error() string
  Code() int
  Cause() error
}

type errorData struct {
  message string
  code int
  cause error
}
func (e errorData) Error() string {
  return e.message
}
func (e errorData) Code() int {
  return e.code
}
func (e errorData) Cause() error {
  return e.cause
}

func annotateError(cause error) error {
  if cause == nil {
    return nil
  }
  // '1' is the 'annotateError' call itself
  // '2' is the error creation point
  pc, fn, line, _ := runtime.Caller(2)
  return errors.New(fmt.Sprintf("(%s[%s:%d]) %s", runtime.FuncForPC(pc).Name(), fn, line, cause))
}


func BadRequestError(message string, cause error) errorData {
  return errorData{message, http.StatusBadRequest, annotateError(cause)}
}
func AuthorizationError(message string, cause error) errorData {
  return errorData{message, http.StatusUnauthorized, annotateError(cause)}
}
func NotFoundError(message string, cause error) errorData {
  return errorData{message, http.StatusNotFound, annotateError(cause)}
}
func UnprocessableEntityError(message string, cause error) errorData {
  return errorData{message, http.StatusUnprocessableEntity, annotateError(cause)}
}
func ServerError(message string, cause error) errorData  {
  return errorData{message, http.StatusInternalServerError, annotateError(cause)}
}
