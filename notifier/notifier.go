// 通知发送
// 目前支持邮件和 alertover
package notifier

import (
	log "github.com/sirupsen/logrus"
)

func init() {
	// 设置日志格式
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
}
