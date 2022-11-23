## Notifier

用于给用户发送邮件，短信等等

#### 发送 Email

```go
// 注册一个EmailNotifier 
// 参数依次为：smtp服务器地址，发送账户，发送账户密码
email, err = notifier.NewEmailNotifier("smtp.exmail.qq.com:25", "service@laily.me", "123")
// 参数为：接收方账户，发送邮件的主题和内容，内容支持html形式
email.SendMail("aa@laily.me;bb@laily.me","title","content")
```

#### 发送 Alertover

Alertover 是一个实现了手机推送的 app，可以将需要推送的内容发送至 alertover 的服务器，然后推送至手机。source 和 receiver 的需要去 www.alertover.com 获取

```go
	source := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	receiver := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	alert := notifier.NewAlertover(s)
	alert.SendAlert(receiver, "title", "content")
```

