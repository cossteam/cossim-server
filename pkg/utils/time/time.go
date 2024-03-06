package time

import "time"

// 返回当前时间的时间戳/毫秒
func Now() int64 {
	return time.Now().UnixNano() / 1e6
}

func IsTimeDifferenceGreaterThanTwoMinutes(time1, time2 int64) bool {
	const twoMinutesInMillis = 2 * 60 * 1000 // 2 minutes in milliseconds

	difference := time1 - time2
	return difference > twoMinutesInMillis || difference < -twoMinutesInMillis
}
