package time

import "time"

// 返回当前时间的时间戳/毫秒
func Now() int64 {
	return time.Now().UnixNano() / 1e6
}
