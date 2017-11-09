package fav

import (
  "errors"
)

type FavoriteMerchantRepository interface {
  save(merchant Merchants, mobileNo string) error
}

type FavoriteMerchantMongoRepository struct {

}

func (fmr FavoriteMerchantMongoRepository) save(merchant Merchants, mobileNo string) error  {
  return errors.New("Favorite Merchant Not implement yet")
}
