package fav

import (
  "errors"
)

type MerchantRepository interface {
  getDetailBy(id string) (Merchants, error)
}

type MerchantMongoRepository struct {}
func (mr MerchantMongoRepository) getDetailBy(id string) (Merchants, error)  {
  return Merchants{}, errors.New("Merchant Not implement yet")
}
