package rest

import (
  "encoding/json"
  "fmt"
  "net/http"
)

func ExtractJson(w http.ResponseWriter, r *http.Request, d interface{}, dDesc string) RestError {
  decoder := json.NewDecoder(r.Body)

  if err := decoder.Decode(d); err != nil {
    restErr := UnprocessableEntityError(fmt.Sprintf("Could not decode payload: %s", dDesc), err)
    HandleError(w, restErr)
    return restErr
  }
  defer r.Body.Close()

  return nil
}
