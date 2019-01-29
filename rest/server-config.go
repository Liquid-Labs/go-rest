package rest

import (
  "log"
  "os"
)

func MustGetenv(k string) string {
  v := os.Getenv(k)
  if v == "" {
    log.Panicf("%s environment variable not set.", k)
  }
  return v
}

var envPurpose = os.Getenv(`NODE_ENV`)
func GetEnvPurpose() string {
  if envPurpose == "" {
    return `test` // TODO: justify this assumption or change to unknown
  } else {
    return envPurpose
  }
}
