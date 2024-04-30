package time

import "time"

// 返回当前时间的时间戳/毫秒
func Now() int64 {
	return time.Now().UnixNano() / 1e6
}

// FormatTimestamp 将时间戳格式化为指定格式的时间字符串
func FormatTimestamp(timestamp int64) string {
	// 将时间戳转换为 time.Time 类型
	t := time.Unix(0, timestamp*int64(time.Millisecond))

	// 格式化时间字符串
	formattedTime := t.Format("2006-01-02 15:04:05")

	return formattedTime
}

func IsTimeDifferenceGreaterThanTwoMinutes(time1, time2 int64) bool {
	const twoMinutesInMillis = 2 * 60 * 1000 // 2 minutes in milliseconds

	difference := time1 - time2
	return difference > twoMinutesInMillis || difference < -twoMinutesInMillis
}
