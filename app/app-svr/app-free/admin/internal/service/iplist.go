package service

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
)

// IPRangeToIPList 获取网段的IP地址列表
// 返回 IP列表+错误信息
// nolint:gomnd
func IPRangeToIPList(Ipaddr string) ([]string, error) {
	ipRangeList := strings.Split(Ipaddr, "/")
	if len(ipRangeList) != 2 {
		return nil, errors.New("ipaddr format error")
	}
	ip := ipRangeList[0]
	mask, err := strconv.Atoi(ipRangeList[1])
	if err != nil {
		return nil, errors.New("mask string to int error")
	}
	var result []string
	if mask > 32 || mask < 0 {
		return nil, errors.New("mask Error: out range")
	}
	maskhead := 0xFFFFFFFF
	for i := 1; i <= 32-mask; i++ {
		maskhead = maskhead << 1
	}
	masktail := 0xFFFFFFFF
	for i := 1; i <= mask; i++ {
		masktail = masktail >> 1
	}
	ipint := ipStringToInt(ip)
	IPintstart := ipint & maskhead
	IPintend := ipint | masktail

	for i := IPintstart; i <= IPintend; i++ {
		result = append(result, ipIntToString(i))
	}
	return result, nil
}

// 将IP字符串转成数值类型
// 返回 数值类型IP
func ipStringToInt(ipString string) int {
	ipSegs := strings.Split(ipString, ".")
	var (
		ipInt int
		pos   uint = 24
	)
	for _, ipSeg := range ipSegs {
		tempInt, _ := strconv.Atoi(ipSeg)
		tempInt = tempInt << pos
		ipInt = ipInt | tempInt
		pos -= 8
	}
	return ipInt
}

// 将IP数值转成字符串类型
// 返回 字符类型IP
// nolint:gomnd
func ipIntToString(ipInt int) string {
	ipSegs := make([]string, 4)
	length := len(ipSegs)
	buffer := bytes.NewBufferString("")
	for i := 0; i < length; i++ {
		tempInt := ipInt & 0xFF
		ipSegs[length-i-1] = strconv.Itoa(tempInt)
		ipInt = ipInt >> 8
	}
	for i := 0; i < length; i++ {
		buffer.WriteString(ipSegs[i])
		if i < length-1 {
			buffer.WriteString(".")
		}
	}
	return buffer.String()
}
