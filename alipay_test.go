package alipay

import (
	"testing"
)

func TestV1(t *testing.T) {
	AlipayPartner = "0"
	AlipayKey = "0"
	WebReturnUrl = "none"
	WebNotifyUrl = "none"
	WebSellerEmail = "huangziyi@wokugame.com"
	form := CreateAlipaySign("123", 99.8, "翱翔大空", "充值100")
	if form == "" {
		t.Error("v1错误")
	}
}

func TestNew(t *testing.T) {
	alipay := Client{
		Partner:   "", // 合作者ID
		Key:       "", // 合作者私钥
		ReturnUrl: "", // 同步返回地址
		NotifyUrl: "", // 网站异步返回地址
		Email:     "", // 网站卖家邮箱地址
	}
	form := alipay.Form(Options{
		OrderId:  "123",
		Fee:      99.8,
		NickName: "翱翔大空",
		Subject:  "充值100",
	})
	if form == "" {
		t.Error("最新接口错误")
	}
}
