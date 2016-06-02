// @authors     ascoders

package alipay

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/astaxie/beego"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Client struct {
	Partner   string // 合作者ID
	Key       string // 合作者私钥
	ReturnUrl string // 同步返回地址
	NotifyUrl string // 网站异步返回地址
	Email     string // 网站卖家邮箱地址
}

type Result struct {
	// 状态
	Status int
	// 本网站订单号
	OrderNo string
	// 支付宝交易号
	TradeNo string
	// 买家支付宝账号
	BuyerEmail string
	// 错误提示
	Message string
}

// 生成订单的参数
type Options struct {
	OrderId  string  // 订单唯一id
	Fee      float32 // 价格
	NickName string  // 充值账户名称
	Subject  string  // 充值描述
}

/* 生成支付宝即时到帐提交表单html代码 */
func (this *Client) Form(opts Options) string {
	//实例化参数
	param := &AlipayParameters{}
	param.InputCharset = "utf-8"
	param.Body = "为" + opts.NickName + "充值" + strconv.FormatFloat(float64(opts.Fee), 'f', 2, 32) + "元"
	param.NotifyUrl = this.NotifyUrl
	param.OutTradeNo = opts.OrderId
	param.Partner = this.Partner
	param.PaymentType = 1
	param.ReturnUrl = this.ReturnUrl
	param.SellerEmail = this.Email
	param.Service = "create_direct_pay_by_user"
	param.Subject = opts.Subject
	param.TotalFee = opts.Fee

	//生成签名
	sign := sign(param)

	//追加参数
	param.Sign = sign
	param.SignType = "MD5"

	//生成自动提交form
	return `
		<form id="alipaysubmit" name="alipaysubmit" action="https://mapi.alipay.com/gateway.do?_input_charset=utf-8" method="get" style='display:none;'>
			<input type="hidden" name="_input_charset" value="` + param.InputCharset + `">
			<input type="hidden" name="body" value="` + param.Body + `">
			<input type="hidden" name="notify_url" value="` + param.NotifyUrl + `">
			<input type="hidden" name="out_trade_no" value="` + param.OutTradeNo + `">
			<input type="hidden" name="partner" value="` + param.Partner + `">
			<input type="hidden" name="payment_type" value="` + strconv.Itoa(int(param.PaymentType)) + `">
			<input type="hidden" name="return_url" value="` + param.ReturnUrl + `">
			<input type="hidden" name="seller_email" value="` + param.SellerEmail + `">
			<input type="hidden" name="service" value="` + param.Service + `">
			<input type="hidden" name="subject" value="` + param.Subject + `">
			<input type="hidden" name="total_fee" value="` + strconv.FormatFloat(float64(param.TotalFee), 'f', 2, 32) + `">
			<input type="hidden" name="sign" value="` + param.Sign + `">
			<input type="hidden" name="sign_type" value="` + param.SignType + `">
		</form>
		<script>
			document.forms['alipaysubmit'].submit();
		</script>
	`
}

/* 被动接收支付宝同步跳转的页面 */
func (this *Client) Return(contro *beego.Controller) *Result {
	// 列举全部传参
	type Params struct {
		Body        string `form:"body" json:"body"`                 // 描述
		BuyerEmail  string `form:"buyer_email" json:"buyer_email"`   // 买家账号
		BuyerId     string `form:"buyer_id" json:"buyer_id"`         // 买家ID
		Exterface   string `form:"exterface" json:"exterface"`       // 接口名称
		IsSuccess   string `form:"is_success" json:"is_success"`     // 交易是否成功
		NotifyId    string `form:"notify_id" json:"notify_id"`       // 通知校验id
		NotifyTime  string `form:"notify_time" json:"notify_time"`   // 校验时间
		NotifyType  string `form:"notify_type" json:"notify_type"`   // 校验类型
		OutTradeNo  string `form:"out_trade_no" json:"out_trade_no"` // 在网站中唯一id
		PaymentType uint8  `form:"payment_type" json:"payment_type"` // 支付类型
		SellerEmail string `form:"seller_email" json:"seller_email"` // 卖家账号
		SellerId    string `form:"seller_id" json:"seller_id"`       // 卖家id
		Subject     string `form:"subject" json:"subject"`           // 商品名称
		TotalFee    string `form:"total_fee" json:"total_fee"`       // 总价
		TradeNo     string `form:"trade_no" json:"trade_no"`         // 支付宝交易号
		TradeStatus string `form:"trade_status" json:"trade_status"` // 交易状态 TRADE_FINISHED或TRADE_SUCCESS表示交易成功
		Sign        string `form:"sign" json:"sign"`                 // 签名
		SignType    string `form:"sign_type" json:"sign_type"`       // 签名类型
	}

	// 实例化参数
	param := &Params{}

	// 结果
	result := &Result{}

	// 解析表单内容，失败返回错误代码-3
	if err := contro.ParseForm(param); err != nil {
		result.Status = -3
		result.Message = "解析表单失败"
		return result
	}
	// 如果最基本的网站交易号为空，返回错误代码-1
	if param.OutTradeNo == "" { //不存在交易号
		result.Status = -1
		result.Message = "站交易号为空"
		return result
	} else {
		// 生成签名
		sign := sign(param)

		// 对比签名是否相同
		if sign == param.Sign { //只有相同才说明该订单成功了
			// 判断订单是否已完成
			if param.TradeStatus == "TRADE_FINISHED" || param.TradeStatus == "TRADE_SUCCESS" { //交易成功
				result.Status = 1
				result.OrderNo = param.OutTradeNo
				result.TradeNo = param.TradeNo
				result.BuyerEmail = param.BuyerEmail
				return result
			} else { // 交易未完成，返回错误代码-4
				result.Status = -4
				result.Message = "交易未完成"
				return result
			}
		} else { // 签名认证失败，返回错误代码-2
			result.Status = -2
			result.Message = "签名认证失败"
			return result
		}
	}

	// 位置错误类型-5
	result.Status = -5
	result.Message = "位置错误"
	return result
}

/* 被动接收支付宝异步通知 */
func (this *Client) Notify(contro *beego.Controller) *Result {
	// 从body里读取参数，用&切割
	postArray := strings.Split(string(contro.Ctx.Input.CopyBody(beego.BConfig.MaxMemory)), "&")

	// 实例化url
	urls := &url.Values{}

	// 保存传参的sign
	var paramSign string
	var sign string

	// 如果字符串中包含sec_id说明是手机端的异步通知
	if strings.Index(string(contro.Ctx.Input.CopyBody(beego.BConfig.MaxMemory)), `alipay.wap.trade.create.direct`) == -1 { // 快捷支付
		for _, v := range postArray {
			detail := strings.Split(v, "=")

			// 使用=切割字符串 去除sign和sign_type
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

		// 获取<trade_status></trade_status>之间的request_token
		re, _ := regexp.Compile("\\<trade_status[\\S\\s]+?\\</trade_status>")
		rt := re.FindAllString(sign, 1)
		trade_status := strings.Replace(rt[0], "<trade_status>", "", -1)
		trade_status = strings.Replace(trade_status, "</trade_status>", "", -1)
		urls.Add("trade_status", trade_status)

		// 获取<out_trade_no></out_trade_no>之间的request_token
		re, _ = regexp.Compile("\\<out_trade_no[\\S\\s]+?\\</out_trade_no>")
		rt = re.FindAllString(sign, 1)
		out_trade_no := strings.Replace(rt[0], "<out_trade_no>", "", -1)
		out_trade_no = strings.Replace(out_trade_no, "</out_trade_no>", "", -1)
		urls.Add("out_trade_no", out_trade_no)

		// 获取<buyer_email></buyer_email>之间的request_token
		re, _ = regexp.Compile("\\<buyer_email[\\S\\s]+?\\</buyer_email>")
		rt = re.FindAllString(sign, 1)
		buyer_email := strings.Replace(rt[0], "<buyer_email>", "", -1)
		buyer_email = strings.Replace(buyer_email, "</buyer_email>", "", -1)
		urls.Add("buyer_email", buyer_email)

		// 获取<trade_no></trade_no>之间的request_token
		re, _ = regexp.Compile("\\<trade_no[\\S\\s]+?\\</trade_no>")
		rt = re.FindAllString(sign, 1)
		trade_no := strings.Replace(rt[0], "<trade_no>", "", -1)
		trade_no = strings.Replace(trade_no, "</trade_no>", "", -1)
		urls.Add("trade_no", trade_no)
	}
	// 追加密钥
	sign += this.Key

	// 返回参数
	result := &Result{}
	
	// md5加密
	m := md5.New()
	m.Write([]byte(sign))
	sign = hex.EncodeToString(m.Sum(nil))
	if paramSign == sign { // 传进的签名等于计算出的签名，说明请求合法
		// 判断订单是否已完成
		if urls.Get("trade_status") == "TRADE_FINISHED" || urls.Get("trade_status") == "TRADE_SUCCESS" { //交易成功
			contro.Ctx.WriteString("success")
			result.Status = 1
			result.OrderNo = urls.Get("out_trade_no")
			result.TradeNo = urls.Get("trade_no")
			result.BuyerEmail = urls.Get("buyer_email")
			return result
		} else {
			contro.Ctx.WriteString("error")
		}
	} else {
		contro.Ctx.WriteString("error")
		// 签名不符，错误代码-1
		result.Status = -1
		result.Message = "签名不符"
		return result
	}
	// 未知错误-2
	result.Status = -2
	result.Message = "未知错误"
	return result
}
