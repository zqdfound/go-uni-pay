package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateOrderNo 生成订单号
func GenerateOrderNo(prefix string) string {
	return fmt.Sprintf("%s%s", prefix, time.Now().Format("20060102150405"))
}

// MD5 计算MD5
func MD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// InArray 判断元素是否在数组中
func InArray(needle string, haystack []string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

// ParseFloat 解析浮点数
func ParseFloat(s string, defaultValue float64) float64 {
	var result float64
	if _, err := fmt.Sscanf(s, "%f", &result); err != nil {
		return defaultValue
	}
	return result
}

// Ternary 三元运算符
func Ternary(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}
