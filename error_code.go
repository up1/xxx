package main

type ErrorCodeModel struct {
	MessageCode        string `json:"MessageCode"`
	MessageDescription string `json:"MessageDescription"`
}

type ErrorCodeModelWithStatusFlag struct {
	MessageCode        string `json:"MessageCode"`
	MessageDescription string `json:"MessageDescription"`
	StatusFlag         bool   `json:"StatusFlag"`
}

type ErrorMessage struct {
	Success    ErrorCodeModel
	General    ErrorCodeModel
	Connection ErrorCodeModel
	Invalid    struct {
		Request      ErrorCodeModel
		DealParsing ErrorCodeModel
		DealNotFound ErrorCodeModel
		DealSoldOut  ErrorCodeModel
		DealSaleEnd  ErrorCodeModel
		Timeout      ErrorCodeModel
	}
	Timeout struct {
		Session ErrorCodeModel
	}
	Mismatch struct {
		RedisData ErrorCodeModel
	}
}

func (em *ErrorMessage) New() {
	em.Success = ErrorCodeModel{"00", ""}
	em.General = ErrorCodeModel{"22", "ขออภัย ระบบไม่สามารถดำเนินการได้ในขณะนี้ กรุณาลองใหม่อีกครั้ง"}
	em.Connection = ErrorCodeModel{"78", "ขออภัย ระบบไม่สามารถดำเนินการได้ในขณะนี้ กรุณาลองใหม่อีกครั้ง"}

	em.Timeout.Session = ErrorCodeModel{"92", "กรุณาล็อกอินเข้าสู่ระบบ"}

	em.Invalid.Request = ErrorCodeModel{"13", "ขออภัย ระบบไม่สามารถดำเนินการได้ในขณะนี้"}
	em.Invalid.DealParsing = ErrorCodeModel{"96", "ขออภัย ระบบไม่สามารถดำเนินการได้ในขณะนี้"}
	em.Invalid.DealNotFound = ErrorCodeModel{"97", "ขออภัย ดีลนี้ไม่มีอยู่ในระบบ"}
	em.Invalid.DealSoldOut = ErrorCodeModel{"98", "ขออภัย ดีลนี้จำหน่ายหมดแล้ว"}
	em.Invalid.DealSaleEnd = ErrorCodeModel{"98", "ขออภัย ดีลนี้หมดช่วงเวลาจำหน่ายแล้ว"}
	em.Invalid.Timeout = ErrorCodeModel{"90", "กรุณาล็อกอินเข้าสู่ระบบ"}

	em.Mismatch.RedisData = ErrorCodeModel{"85", "ท่านทำรายการเกินระยะเวลาที่กำหนด"}
}

var EM ErrorMessage
