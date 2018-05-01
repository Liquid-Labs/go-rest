package rest

import (
  "encoding/json"
  "fmt"
  "net/http"
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
    // TODO: hide error and give reference number
    errMessage := fmt.Sprintf("Could not format response: %v", err)
    StandardServerError(w, errMessage)

    return err
  }
  w.Write(respBody)

  return nil
}

func StandardServerError(w http.ResponseWriter, message string) {
  // we tried returning JSON, and still want to eventually, but this is quicker
  http.Error(w, message, http.StatusInternalServerError)
}

func HandlePost(w http.ResponseWriter, r *http.Request, d interface{}, dDesc string) error {
  decoder := json.NewDecoder(r.Body)

  if err := decoder.Decode(d); err != nil {
    StandardServerError(w, fmt.Sprintf("Could not decode %s payload: %v", dDesc, err))
    return err
  }
  defer r.Body.Close()

  return nil
}
