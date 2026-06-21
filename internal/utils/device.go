package utils

import "strings"

func DetectDevice(ua string) string {
	if ua == "" {
		return "未知设备"
	}
	lower := strings.ToLower(ua)
	if strings.Contains(lower, "android") {
		return "安卓手机"
	}
	if strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad") {
		return "苹果手机"
	}
	if strings.Contains(lower, "windows nt") {
		return "Win电脑"
	}
	if strings.Contains(lower, "macintosh") || strings.Contains(lower, "mac os") {
		return "Mac电脑"
	}
	return "未知设备"
}
