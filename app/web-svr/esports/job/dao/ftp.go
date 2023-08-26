package dao

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"time"

	"go-common/library/log"

	"github.com/jlaffaye/ftp"
)

const (
	_ftpRetry = 3
	_sleep    = 100 * time.Millisecond
)

// Retry . retry one function until no error
func Retry(callback func() error, retry int, sleep time.Duration) (err error) {
	for i := 0; i < retry; i++ {
		if err = callback(); err == nil {
			return
		}
		time.Sleep(sleep)
	}
	return
}

// FileMd5 calculates the local file's md5 and store it in a file
func (d *Dao) FileMd5(localFile string, localMD5File string) (err error) {
	var (
		content []byte
	)
	if content, err = ioutil.ReadFile(localFile); err != nil {
		log.Error("FileMd5 ReadFile localFile(%s) localMD5File(%s) err(%v)", localFile, localMD5File, err)
		return
	}
	md5hash := md5.New()
	if _, err = io.Copy(md5hash, bytes.NewReader(content)); err != nil {
		log.Error("FileMd5 io.Copy localFile(%s) localMD5File(%s) err(%v)", localFile, localMD5File, err)
		return
	}
	md5 := md5hash.Sum(nil)
	fMd5 := hex.EncodeToString(md5[:])
	file, error := os.OpenFile(localMD5File, os.O_RDWR|os.O_CREATE, 0766)
	if error != nil {
		log.Error("FileMd5 os.OpenFile localFile(%s) localMD5File(%s) err(%v)", localFile, localMD5File, err)
		return
	}
	defer file.Close()
	file.WriteString(fMd5)
	return
}

// UploadFile the file to remote frp server and update the md5 file
func (d *Dao) UploadFile(localFile, remoteDir, remoteFileName string) (err error) {
	var (
		ftpInfo    = d.c.Search.FTP
		ftpConnect *ftp.ServerConn
		content    []byte // file's content
		fileSize   int64
	)
	// Dial
	if ftpConnect, err = ftp.DialTimeout(ftpInfo.Host, time.Duration(ftpInfo.Timeout)); err != nil {
		log.Error("UploadFile ftp.DialTimeout Host(%s) err(%v)", ftpInfo.Host, err)
		return
	}
	defer ftpConnect.Quit()
	// Login
	if err = ftpConnect.Login(ftpInfo.User, ftpInfo.Pass); err != nil {
		log.Error("UploadFile ftpConnect.Login User(%s) err(%v)", ftpInfo.User, err)
		return
	}
	defer ftpConnect.Logout()
	// go to ftp server dir
	if err = ftpConnect.ChangeDir(remoteDir); err != nil {
		log.Error("UploadFile ftpConnect.ChangeDir remoteDir(%s) err(%v)", remoteDir, err)
		return
	}
	// Upload the file
	if content, err = ioutil.ReadFile(localFile); err != nil {
		log.Error("UploadFile ioutil.ReadFile localFile(%s) err(%v)", localFile, err)
		return
	}
	data := bytes.NewBuffer(content)
	if err = Retry(func() (err error) {
		return ftpConnect.Stor(remoteFileName, data)
	}, _ftpRetry, _sleep); err != nil {
		log.Error("UploadFile Stor remoteFileName(%s) err(%v)", remoteFileName, err)
		return
	}
	log.Info("File %s is uploaded successfully, size: %d", remoteFileName, fileSize)
	return
}
