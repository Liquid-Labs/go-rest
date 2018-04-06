package rest

import (
  "encoding/json"
  "fmt"
  "net/http"
)

type standardResponse struct {
  Data interface{} `json:"data"`
}

func PackageResponse(w http.ResponseWriter, d interface{}) (*standardResponse, error) {
  resp := standardResponse{data: d}

  if respBody, err := json.Marshal(resp); err != nil {
    // TODO: hide error and give reference number
    http.Error(w, fmt.Sprintf("Could not format response: %v", err), 500)
    return _, err
  }

  return respBody, _
}
