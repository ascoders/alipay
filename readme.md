# beego版支付宝SDK！

### 支付流程介绍

用户在您的网站点击支付按钮后，您的网站需要经历下述操作：

 1.为该用户生成一个订单存入数据库，保存了该订单的**唯一标识符**，它可以是数据库的唯一id；该订单触发用户id；该订单充值用户id；充值金额；订单创建时间。上述是最基本要存储的订单数据。

 2.调用此SDK让用户跳转到支付宝付款页面。

 3.等待用户完成付款，此时如果用户没有关闭付款页面，大约3秒后会跳转到你网站指定的**同步回调页面**，如果用户关闭了网页，支付宝也会多次异步通知你的**异步回调页面**，这一切都是为了告诉你用户完成了付款。

 4.处理订单，在**同步回调页面**和**异步回调页面**调用此SDK，获取该订单ID，在数据库中查出并给相应账号充值（之后发邮件通知等等），一定要注意防止订单**重复充值**，你可以标记订单的active解决此问题。

### SDK调用示例

全局初始化，您需要填写一些支付宝充值必要信息：

	func init() {
		//init alipay params
		alipay.AlipayPartner = **********
		alipay.AlipayKey = **********
		alipay.WebReturnUrl = "http://www.wokugame.com/someurl"	//replace to your owner return url
		alipay.WebNotifyUrl = "http://www.wokugame.com/someurl"	//replace to your owner notify url
		alipay.WebSellerEmail = "huangziyi@wokugame.com"			//replace to your owner alipay email
	}
	
### 如何调用

	//your unique order id
	//pay money
	//account's nickname you pay to
	//pay's description
	form := alipay.CreateAlipaySign("123", 19.8, "翱翔大空", "充值19.8元")

	//inner this html to your page,it will automatically jump to the Alipay payment page
	this.Data["json"] = form
	this.ServerJson()
	
### 如何接收支付宝同步跳转的页面
**注意这里需要解析get请求参数，为了自动获取，请传入beego的`&this.Controller`**

	/* 接收支付宝同步跳转的页面 */
	func (this *ApiController) AlipayReturn() {
		//错误代码(1为成功) 订单id(使用它查询订单) 买家支付宝账号(这个不错) 支付宝id(支付宝账单id)
		status, orderId, buyerEmail, tradeNo := alipay.AlipayReturn(&this.Controller)
		if status == 1 { //付款成功，处理订单
			//处理订单
		}
	}
	
### 如何接收支付宝异步跳转的页面
**注意这里需要解析get请求参数，为了自动获取，请传入beego的`&this.Controller`**

	/* 被动接收支付宝异步通知的页面 */
	func (this *ApiController) AlipayNotify() {
		status, orderId, buyerEmail, TradeNo := alipay.AlipayNotify(&this.Controller)
		if status == 1 { //付款成功，处理订单
			//处理订单
		}
	}
	
### 结语

支付宝流程基本就四步：初始化、构造请求用户付款、同步跳转付款、异步post接收付款请求