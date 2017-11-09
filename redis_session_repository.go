package main

type RedisSession struct {

}

func (rs RedisSession) getMobileNo() (string, error)  {
  return "123456", nil
}
