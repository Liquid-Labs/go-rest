package rest

import (
  "encoding/json"
  "fmt"
  "net/http"
)

func ExtractJson(w http.ResponseWriter, r *http.Request, d interface{}, dDesc string) error {
  decoder := json.NewDecoder(r.Body)

  if err := decoder.Decode(d); err != nil {
    HandleError(w, UnprocessableEntityError(fmt.Sprintf("Could not decode payload: %s", dDesc), err))
    return err
  }
  defer r.Body.Close()

  return nil
}
