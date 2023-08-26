package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup"
	"go-gateway/app/app-svr/app-free/admin/internal/model"
)

const (
	_ispFail   = "fail"
	_ispBGP    = "bgp"
	_ispRegexp = "regexp"
)

func (s *Service) Pcap(ctx context.Context, fileName string, data []byte) (resp string, err error) {
	if len(data) == 0 {
		return
	}
	fileMD5 := fileMD5(data)
	if v, ok := s.pcapResult.Load(fileMD5); ok {
		if resp, ok = v.(string); ok {
			return
		}
	}
	dateStr := time.Now().Format("20060102")
	dirPath := fmt.Sprintf("%s/%s/", "pcap", dateStr)
	_, err = os.Stat(dirPath)
	if os.IsNotExist(err) {
		mask := syscall.Umask(0)
		defer syscall.Umask(mask)
		if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return
		}
	}
	f, err := os.Create(dirPath + fileName)
	if err != nil {
		return
	}
	defer f.Close()
	n, err := f.Write(data)
	if err != nil {
		return
	}
	if n != len(data) {
		return
	}
	if resp, err = s.tshark(dirPath + fileName); err != nil {
		return
	}
	s.pcapResult.Store(fileMD5, resp)
	return
}

// nolint:gomnd
func (s *Service) tshark(fileName string) (data string, err error) {
	if fileName == "" {
		return
	}
	var (
		out   []byte
		hostm map[string]string
	)
	eg := errgroup.Group{}
	eg.Go(func() (err error) {
		// tshark -r "..path.." -T fields -e ip.src -e tcp.srcport -e ip.dst -e tcp.dstport -e frame.len -e http.request.full_uri|awk '{if($5||$6){if($2<10000){r=$1}else{r=$3};s[r]+=$5;c[r]+=1;if($6){l[r]=(l[r]","$6)}}}END{for(i in c){print i,c[i],s[i],l[i]}}'
		out, err = exec.Command("tshark", "-r", fileName, "-T", "fields", "-e", "ip.src", "-e", "tcp.srcport", "-e", "ip.dst", "-e", "tcp.dstport", "-e", "frame.len", "-e", "http.request.full_uri").CombinedOutput()
		return
	})
	eg.Go(func() (err error) {
		hostm, err = hostMap(fileName)
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	lines := bytes.Split(out, []byte("\n"))
	urim := make(map[string][]string, len(lines))
	for _, line := range lines {
		s := strings.TrimSpace(string(line))
		ss := strings.Split(s, " ")
		if len(ss) != 6 {
			continue
		}
		uri := ss[5]
		if strings.HasPrefix(uri, "http://") {
			urim[ss[2]] = append(urim[ss[2]], uri)
			continue
		}
		log.Warn("tshark full_uri analysis data(%s) not have prefix(http://)", uri)
	}
	cmd := exec.Command("awk", `{if($5||$6){if($2<10000){r=$1}else{r=$3};s[r]+=$5;c[r]+=1;if($6){l[r]=(l[r]","$6)}}}END{for(i in c){print i,c[i],s[i],l[i]}}`)
	cmd.Stdin = bytes.NewReader(out)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return
	}
	var (
		records               []*model.TFRecord
		failCount, totalCount int
		failSize, totalSize   int
	)
	lines = bytes.Split(out, []byte("\n"))
	for _, line := range lines {
		str := strings.TrimSpace(string(line))
		ss := strings.Split(str, " ")
		if len(ss) < 3 {
			continue
		}
		remote, countStr, sizeStr := ss[0], ss[1], ss[2]
		res, err := s.locDao.Info(context.Background(), remote)
		if err != nil {
			log.Error("%+v", err)
		}
		info := fmt.Sprintf("%s,%s,%s,%s", res.GetInfo().GetCountry(), res.GetInfo().GetProvince(), res.GetInfo().GetCity(), res.GetInfo().GetIsp())
		count, err := strconv.Atoi(countStr)
		if err != nil {
			log.Error("%+v", err)
		}
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			log.Error("%+v", err)
		}
		var fullURI string
		if len(ss) == 4 {
			uriStr := ss[3]
			uris := strings.Split(uriStr, ",")
			urim := make(map[string]struct{}, len(uris))
			uriList := make([]string, 0, len(uris))
			for _, uri := range uris {
				if uri == "" {
					continue
				}
				if _, ok := urim[uri]; !ok {
					urim[uri] = struct{}{}
					uriList = append(uriList, uri)
				}
			}
			if len(uriList) != 0 {
				fullURI = uriList[0]
			}
		}
		if !model.IsIPv4(remote) {
			continue
		}
		tfISP := s.matchIP(remote)
		host, ok := hostm[remote]
		if !ok {
			host = "-"
		}
		if tfISP == _ispFail {
			tfISP = s.matchHost(host)
		}
		if tfISP == _ispFail {
			failCount += count
			failSize += size
		}
		totalCount += count
		totalSize += size
		record := &model.TFRecord{
			ISP:        tfISP,
			RemoteIP:   remote,
			RemoteHost: host,
			Count:      count,
			Size:       size,
			FullURI:    fullURI,
			Info:       info,
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].ISP < records[j].ISP {
			return true
		}
		if records[i].ISP == records[j].ISP {
			if records[i].Size > records[j].Size {
				return true
			}
		}
		return false
	})
	data = fmt.Sprintf("%-5s %-16s %10s %10s %-35s %-35s %35s\n", "tf", "remote_addr", "count", "size", "remote_host", "addr_info", "full_uri")
	for _, r := range records {
		data += fmt.Sprintf("%-5s %-16s %10d %10d %-35s %-35s %35s\n", r.ISP, r.RemoteIP, r.Count, r.Size, r.RemoteHost, r.Info, r.FullURI)
	}
	data += fmt.Sprintf("%-22s %9.2f%% %9.2f%% %-35s\n", "[FAIL_PERCENT]", float64(failCount)/float64(totalCount)*float64(100), float64(failSize)/float64(totalSize)*float64(100), "")
	return
}

// nolint:gomnd
func hostMap(fileName string) (hostm map[string]string, err error) {
	// tshark -nr "..path.."  -Y dns -V|sed -En 's/: type A, class IN(, addr)?//p'|uniq
	out, err := exec.Command("tshark", "-nr", fileName, "-Y", "dns", "-V").CombinedOutput()
	if err != nil {
		return
	}
	cmd := exec.Command("sed", "-En", "s/: type A, class IN(, addr)?//p")
	cmd.Stdin = bytes.NewReader(out)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return
	}
	cmd = exec.Command("uniq")
	cmd.Stdin = bytes.NewReader(out)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return
	}
	lines := bytes.Split(out, []byte("\n"))
	var host string
	hostm = make(map[string]string, len(lines))
	for _, line := range lines {
		s := strings.TrimSpace(string(line))
		ss := strings.Split(s, " ")
		switch len(ss) {
		case 1:
			host = ss[0]
		case 2:
			hostm[ss[1]] = host
		default:
			log.Warn("pcap hostMap analysis host data(%v) len != 1 or 2", ss)
		}
	}
	return
}

func (s *Service) matchIP(ip string) (tfISP string) {
	ipInt := model.InetAtoN(ip)
	tfISP = _ispFail
LOOP:
	for isp, rs := range s.freeRecords {
		for _, r := range rs {
			if ipInt >= r.IPStartInt && ipInt <= r.IPEndInt {
				tfISP = string(isp)
				if r.IsBGP {
					tfISP = _ispBGP
				}
				break LOOP
			}
		}
	}
	return
}

func (s *Service) matchHost(rawurl string) (tfISP string) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return _ispFail
	}
	matched, _ := regexp.MatchString("([0-9]{1,3}.){3}[0-9]{1,3}/|cn-[a-z0-9]{1,10}-[a-z]{1,10}-(v|live|bcache)-[0-9]{1,10}.bilivideo.com/", u.Host+u.Path)
	if matched {
		return _ispRegexp
	}
	return _ispFail
}

func fileMD5(data []byte) string {
	h := md5.New()
	_, _ = h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
