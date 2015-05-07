# golang版支付宝SDK！

做最好的支付宝sdk，大家的支持就是作者最大的动力！

#### 安装

~~~ go
go get github.com/ascoders/alipay
~~~

#### 初始化


初始化，您需要填写一些支付宝充值必要信息：

~~~ go
import "github.com/ascoders/alipay"

func init() {
	//init alipay params
	alipay.AlipayPartner = 0000000000000000
	alipay.AlipayKey = 000000000000000000000	
	alipay.WebReturnUrl = "http://www.wokugame.com/alipay/return"
	alipay.WebNotifyUrl = "http://www.wokugame.com/alipay/notify"
	alipay.WebSellerEmail = "huangziyi@wokugame.com"
}
~~~
	
#### 生成付款表单

~~~ go
/* @params string		unique order id
 * @params float32	pay money
 * @params string		payment account nickname
 * @params string		pay description
 */
form := alipay.CreateAlipaySign("123", 19.8, "翱翔大空", "充值19.8元")

// render "form"
fmt.Println(form)
~~~
	
以上生成的form放在页面任何位置，会利用js自动跳转到支付宝付款页面。
	
#### 监听支付宝回调页面

注意这里需要解析get请求参数，为了自动获取，请传入beego的`&this.Controller`

~~~ go
func (this *ApiController) Test() {
	//错误代码(1为成功) 订单id(使用它查询订单) 买家支付宝账号(这个不错) 支付宝id(支付宝账单id)
	status, orderId, buyerEmail, tradeNo := alipay.AlipayReturn(&this.Controller)
	if status == 1 { //付款成功，处理订单
		//处理订单
	}
}
~~~

参数解析：

`status` - 错误代码

 - -1  最基本的网站交易号为空
 - -2  签名认证失败
 - -3  解析表单内容，失败返回错误代
 - -4  交易未完成

`orderId` - 订单id

它是订单唯一id，用它来查询交易订单

`buyerEmail` - 买家支付宝账号

`tradeNo` - 支付宝订单号

它是支付宝的订单唯一id，虽然与本站订单id没有关系，但在支付宝查询上会用到

#### 监听支付宝异步post信息 

~~~ go
func (this *ApiController) Test() {
	status, orderId, buyerEmail, TradeNo := alipay.AlipayNotify(&this.Controller)
	if status == 1 { //付款成功，处理订单
		//处理订单
	}
}
~~~
	
#### 支付流程介绍

支付宝流程基本就四步：初始化、构造请求用户付款、同步跳转付款、异步post接收付款请求。

用户在您的网站点击支付按钮后，您的网站需要经历下述操作：

 1.为该用户生成一个订单存入数据库，保存了该订单的**唯一标识符**，它可以是数据库的唯一id；该订单触发用户id；该订单充值用户id；充值金额；订单创建时间。上述是最基本要存储的订单数据。

 2.调用此SDK让用户跳转到支付宝付款页面。

 3.等待用户完成付款，此时如果用户没有关闭付款页面，大约3秒后会跳转到你网站指定的**同步回调页面**，如果用户关闭了网页，支付宝也会多次异步通知你的**异步回调页面**，这一切都是为了告诉你用户完成了付款。

 4.处理订单，在**同步回调页面**和**异步回调页面**调用此SDK，获取该订单ID，在数据库中查出并给相应账号充值（之后发邮件通知等等），一定要注意防止订单**重复充值**，你可以标记订单的active解决此问题。

#### 提示

本着轻量便捷的原则，监听回调事件上深度依赖 beego 框架，将监听事件的代码量降到最低。