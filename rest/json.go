package rest

import (
  "encoding/json"
  "errors"
  "fmt"
  "log"
  "net/http"
  "runtime"
)

type PageInfo struct {
  // 1-based index
  CurrentPage    int64 `json:"currentPage"`
  ItemsPerPage   int64 `json:"itemsPerPage"`
  TotalItemCount int64 `json:"totalItemCount"`
}

type standardResponse struct {
  Data     interface{} `json:"data"`
  Message  string      `json:"message"`
  PageInfo *PageInfo   `json:"pageData",omitempty`
}

func StandardResponse(w http.ResponseWriter, d interface{}, message string, pageInfo *PageInfo) (error) {
  w.Header().Set("Content-Type", "application/json")

  resp := standardResponse{Data: d, Message: message, PageInfo: pageInfo}

  var respBody []byte
  var err error
  if respBody, err = json.Marshal(resp); err != nil {
    restErr := ServerError("Could not format response.", err)
    HandleError(w, restErr)

    return restErr
  }
  w.Write(respBody)

  return nil
}

func HandleError(w http.ResponseWriter, err RestError) {
  // Note that ultimately, we want to encode the error in JSON, but it was
  // proving problematic, so for now it's just text.
  if err.Code() == http.StatusInternalServerError {
    // TODO: hide error and give reference number
    log.Printf("ERROR: %+v", err.Cause()) // Log server/untyped errors.
  }

  http.Error(w, err.Error(), err.Code())
}

func ExtractJson(w http.ResponseWriter, r *http.Request, d interface{}, dDesc string) error {
  decoder := json.NewDecoder(r.Body)

  if err := decoder.Decode(d); err != nil {
    HandleError(w, UnprocessableEntityError(fmt.Sprintf("Could not decode payload: %s", dDesc), err))
    return err
  }
  defer r.Body.Close()

  return nil
}

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
  pc, fn, line, _ := runtime.Caller(2) // '1' is the error creation func
  return errors.New(fmt.Sprintf("(%s[%s:%d]) %s", runtime.FuncForPC(pc).Name(), fn, line, cause))
}
func BadRequestError(message string, cause error) errorData {
  return errorData{message, http.StatusBadRequest, annotateError(cause)}
}
func AuthorizationError(message string, cause error) errorData {
  return errorData{message, http.StatusUnauthorized, annotateError(cause)}
}
func UnprocessableEntityError(message string, cause error) errorData {
  return errorData{message, http.StatusUnprocessableEntity, annotateError(cause)}
}
func ServerError(message string, cause error) errorData  {
  return errorData{message, http.StatusInternalServerError, annotateError(cause)}
}
