package delay

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"go-common/library/log"
)

const (
	_bucket = "api-gateway"
)

// 该方法是四个方法的集合,也可以单独使用里面的方法
// apiName:应用名
// loadPath:需要发布的代码的本地地址
func (d *dao) MergeStep(ctx context.Context, apiName, loadPath string) (err error) {
	info, err := d.GetLatestWF(ctx, apiName)
	if err != nil || info.Version == "" {
		return
	}
	var (
		data     []byte                              //压缩包数据
		url      string                              //boss的完整url
		version  = info.Version                      //版本 应用打包后使用 应用与版本是1:N关系
		fileName = fmt.Sprintf("%s.tar.gz", version) //应用名 以.tar.gz结尾
		destPath = fmt.Sprintf("/tmp/%s", fileName)  //压缩包存放地址
	)
	defer func() {
		if err != nil {
			_ = d.UpdateBoss(ctx, info.ID, url, DisplayNameUploadBoss, DisplayStateFailed)
		}
	}()
	if err = d.Compress(loadPath, destPath); err != nil {
		return
	}
	if data, err = d.ReadTarPackage(destPath); err != nil {
		return
	}
	if url, err = d.Upload(ctx, apiName, fileName, data); err != nil {
		return
	}
	if err = d.UpdateBoss(ctx, info.ID, url, DisplayNameUploadBoss, DisplayStateSucceeded); err != nil {
		return
	}
	return
}

func (d *dao) Upload(ctx context.Context, prefix, filename string, fileData []byte) (res string, err error) {
	path := fmt.Sprintf("%s/%s", prefix, filename)
	if _, err = d.boss.PutObject(ctx, _bucket, path, fileData); err != nil {
		log.Errorc(ctx, "delay upload failed error:%+v", err)
		return
	}
	res = fmt.Sprintf("%s/%s/%s", d.host.Boss, _bucket, path)
	return
}

func (d *dao) ReadTarPackage(destPath string) (res []byte, err error) {
	var fd *os.File
	if fd, err = os.Open(destPath); err != nil {
		log.Error("os.Open failed error:%+v", err)
		return
	}
	if res, err = ioutil.ReadAll(fd); err != nil {
		log.Error("ReadTarPackage failed error:%+v", err)
	}
	return
}

func (d *dao) Compress(filePath string, destPath string) (err error) {
	var (
		fd   *os.File
		dest *os.File
	)
	if fd, err = os.Open(filePath); err != nil {
		return
	}
	defer fd.Close()
	if dest, err = os.Create(destPath); err != nil {
		log.Error("os.Create failed error:%+v", err)
		return
	}
	defer dest.Close()
	gw := gzip.NewWriter(dest)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	if err = compress(fd, "", tw); err != nil {
		log.Error("delay.compress failed error:%+v", err)
		return
	}
	return
}

// https://studygolang.com/articles/7481
func compress(file *os.File, prefix string, tw *tar.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, tw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = prefix + "/" + header.Name
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
