package random

import (
	"math/rand"
	"time"
	"unsafe"
)

/**
 * @Description
 * @Author r0cky
 * @Date 2022/2/15 12:18
 */

const (
	chars    = "0123456789abcdef" //ghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_
	charsLen = len(chars)
	mask     = 1<<6 - 1
)

var rng = rand.NewSource(time.Now().UnixNano())

// RandStr 返回指定长度的随机字符串
func RandStr(ln int) string {
	/* chars 38个字符
	 * rng.Int63() 每次产出64bit的随机数,每次我们使用6bit(2^6=64) 可以使用10次
	 */
	buf := make([]byte, ln)
	for idx, cache, remain := ln-1, rng.Int63(), 10; idx >= 0; {
		if remain == 0 {
			cache, remain = rng.Int63(), 10
		}
		buf[idx] = chars[int(cache&mask)%charsLen]
		cache >>= 6
		remain--
		idx--
	}
	return *(*string)(unsafe.Pointer(&buf))
}
