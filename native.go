// author : jarryliu
// date : 2015-07-28 02:22

package alipay

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type AliPayParameters struct {
	InputCharset string  `json:"_input_charset"` //网站编码
	Body         string  `json:"body"`           //订单描述
	NotifyUrl    string  `json:"notify_url"`     //异步通知页面
	OutTradeNo   string  `json:"out_trade_no"`   //订单唯一id
	Partner      string  `json:"partner"`        //合作者身份ID
	PaymentType  uint8   `json:"payment_type"`   //支付类型 1：商品购买
	ReturnUrl    string  `json:"return_url"`     //回调url
	SellerEmail  string  `json:"seller_email"`   //卖家支付宝邮箱
	Service      string  `json:"service"`        //接口名称
	Subject      string  `json:"subject"`        //商品名称
	TotalFee     float32 `json:"total_fee"`      //总价
	Sign         string  `json:"sign"`           //签名，生成签名时忽略
	SignType     string  `json:"sign_type"`      //签名类型，生成签名时忽略
}

/* 被动接收支付宝同步跳转的页面 */
func (this *Client) NativeReturn(r *http.Request) Result {
	var result Result

	//实例化参数
	param := map[string]string{
		"body":         "", //描述
		"buyer_email":  "", //买家账号
		"buyer_id":     "", //买家ID
		"exterface":    "",
		"is_success":   "", //交易是否成功
		"notify_id":    "", //通知校验id
		"notify_time":  "", //校验时间
		"notify_type":  "", //校验类型
		"out_trade_no": "", //在网站中唯一id
		"payment_type": "", //支付类型
		"seller_email": "", //卖家账号
		"seller_id":    "", //卖家id
		"subject":      "", //商品名称
		"total_fee":    "", //总价
		"trade_no":     "", //支付宝交易号
		"trade_status": "", //交易状态 TRADE_FINISHED或TRADE_SUCCESS表示交易成功
		"sign":         "", //签名
		"sign_type":    "", //签名类型
	}

	//解析表单内容，失败返回错误代码-3
	form := r.URL.Query()
	for k, _ := range param {
		param[k] = form.Get(k)
	}

	result.OrderNo = param["out_trade_no"]
	result.TradeNo = param["trade_no"]

	//如果最基本的网站交易号为空，返回错误代码-1
	if result.OrderNo == "" { //不存在交易号
		result.Status = -1
		return result
	}
	//生成签名
	sign := sign(param)
	//对比签名是否相同
	if sign == param["sign"] { //只有相同才说明该订单成功了
		//判断订单是否已完成
		tradeStatus := param["trade_status"]
		if tradeStatus == "TRADE_FINISHED" || tradeStatus == "TRADE_SUCCESS" { //交易成功
			result.Status = 1
		} else { //交易未完成，返回错误代码-4
			result.Status = -4
		}
	} else { //签名认证失败，返回错误代码-2
		result.Status = -2
	}

	//位置错误类型-5
	if result.Status == 0 {
		result.Status = -5
	}

	return result
}

/* 被动接收支付宝异步通知 */
func (this *Client) NativeNotify(r *http.Request) Result {

	// /pay/notify/104_alipay?discount=0.00&payment_type=1&subject=%E5%9C%A8%E7%BA%BF%E6%94%AF%E4%BB%98%E8%AE%A2%E5%8D%95&trade_no=2015072800001000810060741985&buyer_email=***&gmt_create=2015-07-28%2001:24:19%C2%ACify_type=trade_status_sync&quantity=1&out_trade_no=146842585&seller_id=2088021187655650%C2%ACify_time=2015-07-28%2001:24:29&body=%E8%AE%A2%E5%8D%95%E5%8F%B7%EF%BC%9A146842585&trade_status=TRADE_SUCCESS&is_total_fee_adjust=N&total_fee=0.01&gmt_payment=2015-07-28%2001:24:29&seller_email=***&price=0.01&buyer_id=2088302384317810%C2%ACify_id=75e570fcc802c637d8cf1fdaa8677d046i&use_coupon=N&sign_type=MD5&sign=***
	var result Result
	body, _ := ioutil.ReadAll(r.Body)
	bodyStr := string(body)

	if bodyStr == "" {
		/*
			if len(r.URL.RawPath) != 0 {
				bodyStr = r.URL.RawQuery[1:]
			}
		*/
		result.Status = -4
		return result
	}

	//从body里读取参数，用&切割
	postArray := strings.Split(bodyStr, "&")

	//实例化url
	urls := &url.Values{}

	//保存传参的sign
	var paramSign string
	var sign string

	//如果字符串中包含sec_id说明是手机端的异步通知
	if strings.Index(bodyStr, `alipay.wap.trade.create.direct`) == -1 { //快捷支付
		for _, v := range postArray {
			detail := strings.Split(v, "=")

			//使用=切割字符串 去除sign和sign_type
			if detail[0] == "sign" || detail[0] == "sign_type" {
				if detail[0] == "sign" {
					paramSign = detail[1]
				}
				continue
			} else {
				urls.Add(detail[0], detail[1])
			}
		}

		// url解码
		urlDecode, _ := url.QueryUnescape(urls.Encode())
		sign, _ = url.QueryUnescape(urlDecode)
	} else { // 手机网页支付
		// 手机字符串加密顺序
		mobileOrder := []string{"service", "v", "sec_id", "notify_data"}
		for _, v := range mobileOrder {
			for _, value := range postArray {
				detail := strings.Split(value, "=")
				// 保存sign
				if detail[0] == "sign" {
					paramSign = detail[1]
				} else {
					// 如果满足当前v
					if detail[0] == v {
						if sign == "" {
							sign = detail[0] + "=" + detail[1]
						} else {
							sign += "&" + detail[0] + "=" + detail[1]
						}
					}
				}
			}
		}
		sign, _ = url.QueryUnescape(sign)

		//获取<trade_status></trade_status>之间的request_token
		re, _ := regexp.Compile("\\<trade_status[\\S\\s]+?\\</trade_status>")
		rt := re.FindAllString(sign, 1)
		trade_status := strings.Replace(rt[0], "<trade_status>", "", -1)
		trade_status = strings.Replace(trade_status, "</trade_status>", "", -1)
		urls.Add("trade_status", trade_status)

		//获取<out_trade_no></out_trade_no>之间的request_token
		re, _ = regexp.Compile("\\<out_trade_no[\\S\\s]+?\\</out_trade_no>")
		rt = re.FindAllString(sign, 1)
		out_trade_no := strings.Replace(rt[0], "<out_trade_no>", "", -1)
		out_trade_no = strings.Replace(out_trade_no, "</out_trade_no>", "", -1)
		urls.Add("out_trade_no", out_trade_no)

		//获取<buyer_email></buyer_email>之间的request_token
		re, _ = regexp.Compile("\\<buyer_email[\\S\\s]+?\\</buyer_email>")
		rt = re.FindAllString(sign, 1)
		buyer_email := strings.Replace(rt[0], "<buyer_email>", "", -1)
		buyer_email = strings.Replace(buyer_email, "</buyer_email>", "", -1)
		urls.Add("buyer_email", buyer_email)

		//获取<trade_no></trade_no>之间的request_token
		re, _ = regexp.Compile("\\<trade_no[\\S\\s]+?\\</trade_no>")
		rt = re.FindAllString(sign, 1)
		trade_no := strings.Replace(rt[0], "<trade_no>", "", -1)
		trade_no = strings.Replace(trade_no, "</trade_no>", "", -1)
		urls.Add("trade_no", trade_no)
	}
	//追加密钥
	sign += this.Key

	//md5加密
	m := md5.New()
	m.Write([]byte(sign))
	sign = hex.EncodeToString(m.Sum(nil))

	result.OrderNo = urls.Get("out_trade_no")
	result.TradeNo = urls.Get("trade_no")
	//fee ,_ := strconv.ParseFloat(urls.Get("total_fee"),32)
	//payResult.Fee = float32(fee)

	if paramSign == sign { //传进的签名等于计算出的签名，说明请求合法
		//判断订单是否已完成
		if urls.Get("trade_status") == "TRADE_FINISHED" || urls.Get("trade_status") == "TRADE_SUCCESS" { //交易成功
			result.Status = 1
		}
	} else {
		//签名不符，错误代码-1
		result.Status = -1
	}

	return result
}
