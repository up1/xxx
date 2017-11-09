package fav

import (
  "testing"
  "errors"
)

type MockSession struct {}
func (ms MockSession) getMobileNo() (string, error)  {
  return "123456", nil
}

type MockMerchantRepository struct {}
func (mr MockMerchantRepository) getDetailBy(id string) (Merchants, error)  {
  return Merchants{}, nil
}

type MockFavoriteMerchantRepository struct {}
func (fmr MockFavoriteMerchantRepository) save(merchant Merchants, mobileNo string) error  {
  return nil
}

func TestSucess_To_add_new_shop_to_favorite(t *testing.T)  {
  service := NewFMServiceWithDependencies(
       MockSession{},
       MockMerchantRepository{},
       MockFavoriteMerchantRepository{})
  err := service.add("1234")
  if err != nil {
    t.Errorf("Sad => %v", err)
  }
}


type MockSessionFail struct {}
func (ms MockSessionFail) getMobileNo() (string, error)  {
  return "", errors.New("Can't get data from session")
}

type MockMerchantRepositoryFail struct {}
func (mr MockMerchantRepositoryFail) getDetailBy(id string) (Merchants, error)  {
  return Merchants{}, errors.New("Can't get merchant data")
}

type MockFavoriteMerchantRepositoryFail struct {}
func (fmr MockFavoriteMerchantRepositoryFail) save(merchant Merchants, mobileNo string) error  {
  return errors.New("Can't get favorite merchant data")
}

func TestFailure_no_session(t *testing.T)  {
  service := NewFMServiceWithDependencies(MockSessionFail{}, nil, nil)
  err := service.add("1234")
  if err == nil {
    t.Errorf("Sad no error waaa")
  }
}

func TestFailure_cannot_get_merchant_by_id(t *testing.T)  {
  service := NewFMServiceWithDependencies(MockSession{}, MockMerchantRepositoryFail{}, nil)
  err := service.add("1234")
  if err == nil {
    t.Errorf("Sad no error waaa")
  }
}

func TestFailure_cannot_get_farite_merchant(t *testing.T)  {
  service := NewFMServiceWithDependencies(MockSession{}, MockMerchantRepository{}, MockFavoriteMerchantRepositoryFail{})
  err := service.add("1234")
  if err == nil {
    t.Errorf("Sad no error waaa")
  }
}
