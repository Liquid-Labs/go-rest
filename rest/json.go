package rest

import (
  "encoding/json"
  "fmt"
  "log"
  "net/http"

  "google.golang.org/appengine"
)

type standardResponse struct {
  Data         interface{}   `json:"data"`
  Message      string        `json:"message"`
  SearchParams *SearchParams `json:"searchParams",omitempty`
}

func StandardResponse(w http.ResponseWriter, d interface{}, message string, searchParams *SearchParams) (error) {
  w.Header().Set("Content-Type", "application/json")

  resp := standardResponse{Data: d, Message: message, SearchParams: searchParams}

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

func HandleError(w http.ResponseWriter, err RestError) (RestError) {
  // Note that ultimately, we want to encode the error in JSON, but it was
  // proving problematic, so for now it's just text.
  if err.Code() == http.StatusInternalServerError {
    // TODO: hide error and give reference number
    log.Printf("ERROR: %+v", err.Cause()) // Log server/untyped errors.
  } else if appengine.IsDevAppServer() {
    log.Printf("%+v", err)
    log.Printf("%+v", err.Cause())
  }

  http.Error(w, err.Error(), err.Code())

  return err
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
