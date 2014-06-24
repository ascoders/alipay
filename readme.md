#beego版支付宝SDK！

###支付流程介绍

用户在您的网站点击支付按钮后，您的网站需要经历下述操作：

 1.为该用户生成一个订单存入数据库，保存了该订单的**唯一标识符**，它可以是数据库的唯一id；该订单触发用户id；该订单充值用户id；充值金额；订单创建时间。上述是最基本要存储的订单数据。

 2.调用此SDK让用户跳转到支付宝付款页面。

 3.等待用户完成付款，此时如果用户没有关闭付款页面，大约3秒后会跳转到你网站指定的**同步回调页面**，如果用户关闭了网页，支付宝也会多次异步通知你的**异步回调页面**，这一切都是为了告诉你用户完成了付款。

 4.处理订单，在**同步回调页面**和**异步回调页面**调用此SDK，获取该订单ID，在数据库中查出并给相应账号充值（之后发邮件通知等等），一定要注意防止订单**重复充值**，你可以标记订单的active解决此问题。

###SDK调用示例

全局初始化，您需要填写一些支付宝充值必要信息：

    func init() {
	  //初始化支付宝插件
	  alipay.AlipayPartner = **********
	  alipay.AlipayKey = **********
	  alipay.WebReturnUrl = "http://www.ascode.net/someurl" //替换成你的 异步回调页面
	  alipay.WebNotifyUrl = "http://www.ascode.net/someurl" //替换成你的 同步回调页面
	  alipay.WebSellerEmail = "huangziyi@ascode.net"        //替换成你的 支付宝账号邮箱
    }
	
用户点击支付按钮后如何调用SDK

	//创建订单order，生成了各种信息包括订单的唯一id
	//获取支付宝即时到帐的自动提交表单
	//四个参数分别是 订单唯一id(string) 充值金额(int) 实际充值的游戏币(int) 充值时给用户的描述(string)
	form := alipay.CreateAlipaySign(order.Id.Hex(), int(number), order.Gain, order.AccountPay, "我酷游戏-充值代金券"+strconv.Itoa(order.Gain))
	//为了更好的用户体验，可以以json方式调用，返回了json类型字符串
	this.Data["json"] = form
	
	//前台接收到字符串后直接输出即可跳转
	document.write(data);