package rest

import (
  "encoding/json"
  "fmt"
  "net/http"
)

type standardResponse struct {
  Data    interface{} `json:"data"`
  Message string      `json:"message"`
}

func StandardResponse(w http.ResponseWriter, d interface{}, message string) (error) {
  w.Header().Set("Content-Type", "application/json")

  resp := standardResponse{Data: d, Message: message}

  var respBody []byte
  var err error
  if respBody, err = json.Marshal(resp); err != nil {
    // TODO: hide error and give reference number
    errMessage := fmt.Sprintf("Could not format response: %v", err);
    http.Error(w, message, 500)
    // We will tyr and provide a message in the JSON
    resp.Data = nil
    resp.Message = errMessage;
    if respBody, err = json.Marshal(resp); err == nil {
      w.Write(respBody)
    }
    return err
  }
  w.Write(respBody)

  return nil
}

func StandardServerError(w http.ResponseWriter, message string) {
  w.Header().Set("Content-Type", "application/json")
  // TODO: hide error and give reference number
  resp := standardResponse{Data: nil, Message: message}
  if respBody, err := json.Marshal(resp); err != nil {
    w.Write(respBody)
  }
  http.Error(w, message, 500)
}

func HandlePost(w http.ResponseWriter, r *http.Request, d interface{}, dDesc string) {
  decoder := json.NewDecoder(r.Body)

  if err := decoder.Decode(d); err != nil {
    StandardServerError(w, fmt.Sprintf("Could not decode %s payload: %v", dDesc, err))
    return
  }
  defer r.Body.Close()
}
