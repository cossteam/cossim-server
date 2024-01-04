package utils

import "time"

func ParseDurationFromString(durationString string) (time.Duration, error) {
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		return 0, err
	}
	return duration, nil
}
