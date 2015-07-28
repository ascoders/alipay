# golang版支付宝SDK！

做最好的支付宝sdk，大家的支持就是作者最大的动力！

## 安装

~~~ go
go get github.com/ascoders/alipay
~~~

## 初始化

~~~ go
alipay := alipay.Client{
	Partner   : "", // 合作者ID
	Key       : "", // 合作者私钥
	ReturnUrl : "", // 同步返回地址
	NotifyUrl : "", // 网站异步返回地址
	Email     : "", // 网站卖家邮箱地址
}
~~~

alipay升级到2.0版本，[api_v1.0](doc/v1.md)依旧兼容

## 生成提交表单

~~~ go
form := alipay.Form(alipay.Options{
	OrderId:  "123",	// 唯一订单号
	Fee:      99.8,		// 价格
	NickName: "翱翔大空",	// 用户昵称，支付页面显示用
	Subject:  "充值100",	// 支付描述，支付页面显示用
})
~~~

将form输出到页面上，会自动跳转至支付宝收银台。

## 回调处理

回调分为`同步`和`异步`，支付成功后跳转至商户页面成为`同步回调`，跳转地址在`ReturnUrl`参数中配置。支付宝每隔10 10 30 ... 秒发送一次异步请求称之为`异步回调`，通知地址在`NotifyUrl`中配置。

#### 同步回调（依赖beego）

注意这里需要解析get请求参数，为了自动获取，请传入beego的`&this.Controller`

~~~ go
func (this *ApiController) Return() {
	result := alipay.Return(&this.Controller)
	if result.Status == 1 { //付款成功，处理订单
		//处理订单
	}
}
~~~

参数：

- Partner	// 合作者ID
- Key       // 合作者私钥
- ReturnUrl // 同步返回地址
- NotifyUrl // 网站异步返回地址
- Email     // 网站卖家邮箱地址

至于为什么没有给充值金额参数，因为金额不能代表问题，例如打折或者做活动，请自行查询订单表，根据业务逻辑进行处理。

#### 同步回调（无依赖）

感谢@atnet提供的代码，让alipay支持原生http请求，脱离推框架的依赖。

~~~ go
func ReturnHandle(w http.ResponseWriter, r *http.Request) {
	result := alipay.NativeReturn(r)
	if result.Status == 1 { //付款成功，处理订单
		//处理订单
	}
}
~~~

#### 异步回调（依赖beego）

~~~ go
func (this *ApiController) Notify() {
	result := alipay.Notify(&this.Controller)
	if result.Status == 1 { //付款成功，处理订单
		//处理订单
	}
}
~~~

> 异步回调的返回参数与同步回调相同。

#### 异步回调（无依赖）
	
~~~ go
func NotifyHandle(w http.ResponseWriter, r *http.Request) {
	result := alipay.NativeNotify(r)
	if result.Status == 1 { //付款成功，处理订单
		//处理订单
	}
}
~~~
	
## 支付流程介绍

支付宝流程基本就四步：初始化、构造请求用户付款、同步跳转付款、异步post接收付款请求。

用户在您的网站点击支付按钮后，您的网站需要经历下述操作：

 1.为该用户生成一个订单存入数据库，保存了该订单的**唯一标识符**，它可以是数据库的唯一id；该订单触发用户id；该订单充值用户id；充值金额；订单创建时间。上述是最基本要存储的订单数据。

 2.调用此SDK让用户跳转到支付宝付款页面。

 3.等待用户完成付款，此时如果用户没有关闭付款页面，大约3秒后会跳转到你网站指定的**同步回调页面**，如果用户关闭了网页，支付宝也会多次异步通知你的**异步回调页面**，这一切都是为了告诉你用户完成了付款。

 4.处理订单，在**同步回调页面**和**异步回调页面**调用此SDK，获取该订单ID，在数据库中查出并给相应账号充值（之后发邮件通知等等），一定要注意防止订单**重复充值**，你可以标记订单的active解决此问题。