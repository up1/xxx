package main

type FMService struct {
}

func NewFMService() FMService {
  return FMService{}
}

func (fm FMService) add(merchantId string) error  {
  // 1. GetSessionDataRedis => MOBILE_NO
  session := RedisSession{}
  mobileNo, err := session.getMobileNo()
  if err != nil {
    return err
  }
  // 2. Get merchant info by merchantId
  mr := MerchantRepository{}
  merchant, err := mr.getDetailBy(merchantId)
  if err != nil {
    return err
  }
  // 3. บันทึก
  fmr := FavoriteMerchantRepository{}
  err = fmr.save(merchant, mobileNo)
  return err
}
