package fav

import (
  "errors"
)

type SessionStorage interface {
  getMobileNo() (string, error)
}

type RedisSession struct {}

func (rs RedisSession) getMobileNo() (string, error)  {
  return "", errors.New("Redis Not implement yet")
}
