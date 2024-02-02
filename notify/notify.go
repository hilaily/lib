// 通知发送
// 目前支持邮件和 alertover
package notify

import (
	log "github.com/sirupsen/logrus"
)

func Init() {
	// 设置日志格式
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
}
