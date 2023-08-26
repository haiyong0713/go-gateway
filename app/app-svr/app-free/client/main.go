package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	_ "go.uber.org/automaxprocs"
)

const (
	_addRecordURI = "https://portal.bilibili.co/x/free/external/record/add"
)

var (
	filePath string
)

func init() {
	flag.StringVar(&filePath, "f", "", "default file path")
}

type FreeRecord struct {
	IPStart  string `json:"ip_start"`
	IPEnd    string `json:"ip_end"`
	ISP      string `json:"isp"`
	IsBGP    bool   `json:"is_bgp"`
	Business string `json:"business"`
	State    int    `json:"state"`
}

func main() {
	flag.Parse()
	lines, err := readByLine(filePath)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	rs := parseLines(lines)
	msg, err := post(rs)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	fmt.Println("免流备案录入结果:", msg)
}

func readByLine(filename string) (map[string][]string, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	res := map[string][]string{}
	rows := f.GetRows("全量免流IP库-移动")
	for _, row := range rows {
		if row[0] != "" {
			res["cm"] = append(res["cm"], row[0])
		}
	}
	rows = f.GetRows("全量免流IP库-联通")
	for _, row := range rows {
		if row[0] != "" && row[0] != "IPv4" {
			res["cu"] = append(res["cu"], row[0])
		}
		if row[1] != "" && row[1] != "IPv6" {
			res["cu"] = append(res["cu"], row[1])
		}
	}
	rows = f.GetRows("全量免流IP库-电信")
	for _, row := range rows {
		if (row[0] != "" && row[0] != "目的IP地址") && (row[1] != "" && row[1] != "Mask") {
			res["ct"] = append(res["ct"], row[0]+","+row[1])
		}
	}
	return res, nil
}

// nolint:gomnd
func parseLines(lines map[string][]string) []*FreeRecord {
	var rs []*FreeRecord
	for isp, val := range lines {
		for _, line := range val {
			var (
				ipStart, ipEnd string
				err            error
			)
			line = strings.Replace(string(line), "\n", "", -1)
			line = strings.Replace(string(line), "\r", "", -1)
			line = strings.Replace(string(line), "\r\n", "", -1)
			fields := strings.Split(string(line), ",")
			switch len(fields) {
			case 1:
				cidr := fields[0]
				if ipStart, ipEnd, err = rangeCIDR(cidr); err != nil {
					fmt.Printf("%+v\n", err)
					continue
				}
			case 2:
				ip, mask := fields[0], fields[1]
				if ipStart, ipEnd, err = rangeIPMask(ip, mask); err != nil {
					fmt.Printf("%+v\n", err)
					continue
				}
			}
			if ipStart == "" || ipEnd == "" {
				continue
			}
			r := &FreeRecord{
				IPStart: ipStart,
				IPEnd:   ipEnd,
				ISP:     isp,
				State:   1,
			}
			rs = append(rs, r)
		}
	}
	return rs
}

func rangeIPMask(ip, mask string) (ipStart, ipEnd string, err error) {
	stringMask := net.IPMask(net.ParseIP(mask).To4())
	ones, _ := stringMask.Size()
	cidr := fmt.Sprintf("%s/%d", ip, ones)
	return rangeCIDR(cidr)
}

func rangeCIDR(cidr string) (ipStart, ipEnd string, err error) {
	if !strings.Contains(cidr, "/") {
		ipStart, ipEnd = cidr, cidr
		return
	}
	if strings.Contains(cidr, "/32") {
		cidr = strings.Replace(cidr, "/32", "", -1)
		ipStart, ipEnd = cidr, cidr
		return
	}
	ips, err := hosts(cidr)
	if err != nil {
		return
	}
	if len(ips) == 0 {
		return
	}
	ipStart = ips[0]
	ipEnd = ips[len(ips)-1]
	return
}

func hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	if ip = ip.To4(); ip == nil {
		return nil, nil
	}
	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	return ips, nil
}

// http://play.golang.org/p/m8TNTtygK0
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func post(rs []*FreeRecord) (msg string, err error) {
	b, err := json.Marshal(rs)
	if err != nil {
		return
	}
	params := url.Values{}
	params.Set("records", string(b))
	resp, err := http.PostForm(_addRecordURI, params)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	if err = json.Unmarshal(body, &res); err != nil {
		return "", err
	}
	if res.Code == 0 {
		return "成功", nil
	}
	return res.Msg, nil
}
