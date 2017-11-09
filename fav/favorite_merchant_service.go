package fav

type FMService struct {
  session SessionStorage
  mr MerchantRepository
  fmr FavoriteMerchantRepository
}

func NewFMService() FMService {
  return FMService{session: RedisSession{},
                   mr: MerchantMongoRepository{}}
}

func NewFMServiceWithDependencies(session SessionStorage,
  mr MerchantRepository,
  fmr FavoriteMerchantRepository) FMService {
  return FMService{session: session, mr: mr, fmr: fmr}
}

func (fm FMService) add(merchantId string) error  {
  // 1. GetSessionDataRedis => MOBILE_NO
  mobileNo, err := fm.session.getMobileNo()
  if err != nil {
    return err
  }
  // 2. Get merchant info by merchantId
  merchant, err := fm.mr.getDetailBy(merchantId)
  if err != nil {
    return err
  }
  // 3. บันทึก
  err = fm.fmr.save(merchant, mobileNo)
  return err
}
