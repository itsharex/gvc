package vctrl

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/moqsien/goutils/pkgs/archiver"
	"github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/koanfer"
	"github.com/moqsien/goutils/pkgs/request"
	"github.com/moqsien/goutils/pkgs/storage"
	config "github.com/moqsien/gvc/pkgs/confs"
	"github.com/moqsien/gvc/pkgs/utils"
)

type RepoType string

type EncryptoType string

const (
	RepoTypeGithub  RepoType     = "github"
	RepoTypeGitee   RepoType     = "gitee"
	RepoName        string       = "gvc_configs"
	StorageConfName string       = ".remote_storage.json"
	EncryptByAES    EncryptoType = "aes"
	EncryptByZip    EncryptoType = "zip"
	EncryptByNone   EncryptoType = "none"
)

/*
New configuration for gvc using github/gitee as remote storage.
*/
type StorageConf struct {
	Type        RepoType `json,koanf:"type"`
	UserName    string   `json,koanf:"username"` // username for github or gitee.
	AccessToken string   `json,koanf:"token"`
	CryptoKey   string   `json,koanf:"crypto_key"` // Key to encrypt your private data like passwords and ssh files, etc.
	ProxyURI    string   `json,koanf:"proxy_uri"`
}

type Synchronizer struct {
	CNF     *StorageConf
	storage storage.IStorage
	path    string
	koanfer *koanfer.JsonKoanfer
}

func NewSynchronizer() (s *Synchronizer) {
	s = &Synchronizer{}
	s.path = filepath.Join(config.GetGVCWorkDir(), StorageConfName)
	s.koanfer, _ = koanfer.NewKoanfer(s.path)
	s.initiate()
	return
}

func (that *Synchronizer) initiate() {
	if that.koanfer == nil {
		gprint.PrintError("nil koanfer.")
		return
	}
	// TODO: use bubbletea TUI.
	if that.CNF.AccessToken == "" {
		var SType RepoType
		fmt.Println("Choose your repo type: ")
		fmt.Println("1. Github. ")
		fmt.Println("2. Gitee. ")
		fmt.Scanln(&SType)

		var username string
		fmt.Println("Enter your username: ")
		fmt.Scanln(&username)

		var token string
		fmt.Println("Enter your access token: ")
		fmt.Scanln(&token)

		var key string
		fmt.Println("Enter your crypto key: ")
		fmt.Scanln(&key)

		var proxyUri string
		fmt.Println("Enter your proxy uri: ")
		fmt.Scanln(&proxyUri)

		that.CNF.Type = SType
		that.CNF.UserName = username
		that.CNF.AccessToken = token
		that.CNF.CryptoKey = key
		that.CNF.ProxyURI = proxyUri
		that.koanfer.Save(that.CNF)
	}

	that.koanfer.Load(that.CNF)

	// setup for remote storage.
	switch that.CNF.Type {
	case RepoTypeGithub:
		gh := storage.NewGhStorage(that.CNF.UserName, that.CNF.AccessToken)
		gh.Proxy = that.CNF.ProxyURI
		that.storage = gh
	case RepoTypeGitee:
		that.storage = storage.NewGtStorage(that.CNF.UserName, that.CNF.AccessToken)
	default:
		// TODO: copy file to backupdir.
	}
}

func (that *Synchronizer) upload(fPath, remoteFileName string) (r []byte) {
	if that.storage == nil {
		gprint.PrintError("No remote storages found.")
		return
	}
	content := that.storage.GetContents(RepoName, "", remoteFileName)
	shaStr := gjson.New(content).Get("sha").String()
	return that.storage.UploadFile(RepoName, "", fPath, shaStr)
}

func (that *Synchronizer) UploadFile(fPath, remoteFileName string, et EncryptoType) {
	if ok, _ := utils.PathIsExist(fPath); !ok {
		gprint.PrintError("File not exist: %s", fPath)
		return
	}
	switch et {
	case EncryptByAES:
		cc := crypt.NewCrptWithKey([]byte(that.CNF.CryptoKey))
		content, err := os.ReadFile(fPath)
		if err != nil {
			gprint.PrintError("Read file error: %+v", err)
			return
		}
		if r, err := cc.AesEncrypt([]byte(content)); err != nil {
			gprint.PrintError("Encrypt file error: %+v", err)
			return
		} else {
			fPath = filepath.Join(config.GVCBackupDir, remoteFileName)
			if err = os.WriteFile(fPath, r, os.ModePerm); err != nil {
				gprint.PrintError("Write file error: %+v", err)
				return
			}
		}
	case EncryptByZip:
		if archive, err := archiver.NewArchiver(fPath, config.GVCBackupDir, false); err == nil {
			archive.SetZipName(remoteFileName)
			archive.SetPassword(that.CNF.CryptoKey)
			err = archive.ZipDir()
			if err != nil {
				gprint.PrintError("Zip dir error: %+v", err)
				return
			}
		}
		fPath = filepath.Join(config.GVCBackupDir, remoteFileName)
	default:
		// copy local file to backupdir then upload.
		content, err := os.ReadFile(fPath)
		if err != nil {
			gprint.PrintError("Read file error: %+v", err)
			return
		}
		fPath = filepath.Join(config.GVCBackupDir, remoteFileName)
		if err = os.WriteFile(fPath, content, os.ModePerm); err != nil {
			gprint.PrintError("Write file error: %+v", err)
			return
		}
	}

	r := that.upload(fPath, remoteFileName)
	j := gjson.New(r)
	if j.Get("content.path").String() != "" && j.Get("content.sha").String() != "" {
		gprint.PrintSuccess("uploaded successfully: %s", fPath)
	} else {
		gprint.PrintWarning("error occurred: %s", string(r))
	}
}

func (that *Synchronizer) download(remoteFileName string) (dUrl string) {
	if that.storage == nil {
		gprint.PrintError("No remote storages found.")
		return
	}
	content := that.storage.GetContents(RepoName, "", remoteFileName)
	dUrl = gjson.New(content).Get("download_url").String()
	if dUrl == "" {
		gprint.PrintWarning("can not find %s in remote repo. %s", remoteFileName, string(content))
	}
	return
}

func (that *Synchronizer) DownloadFile(fPath, remoteFileName string, et EncryptoType) {
	dUrl := that.download(remoteFileName)
	if dUrl == "" {
		return
	}
	// download and deploy files.
	fetcher := request.NewFetcher()
	fetcher.Timeout = time.Minute * 30
	fetcher.SetUrl(dUrl)
	fetcher.Proxy = that.CNF.ProxyURI

	srcPath := filepath.Join(config.GVCBackupDir, remoteFileName)
	if size := fetcher.GetAndSaveFile(srcPath, true); size > 20 {
		switch et {
		case EncryptByAES:
			cc := crypt.NewCrptWithKey([]byte(that.CNF.CryptoKey))
			content, err := os.ReadFile(srcPath)
			if err != nil {
				gprint.PrintError("Read file failed: %+v", err)
				return
			}
			if r, err := cc.AesDecrypt([]byte(content)); err != nil {
				gprint.PrintError("Decrypt file failed: %+v", err)
				return
			} else {
				// deploy remote file to local.
				if err = os.WriteFile(fPath, r, os.ModePerm); err != nil {
					gprint.PrintError("Write file failed: %+v", err)
					return
				}
			}
		case EncryptByZip:
			dstDir := filepath.Dir(fPath)
			if archive, err := archiver.NewArchiver(srcPath, dstDir, false); err == nil {
				archive.SetPassword(that.CNF.CryptoKey)
				_, err = archive.UnArchive()
				if err != nil {
					gprint.PrintError("unarchive failed: %+v", err)
					return
				}
				gprint.PrintSuccess("download successed: %s", fPath)
			}
		default:
			content, err := os.ReadFile(srcPath)
			if err != nil {
				gprint.PrintError("Read file failed: %+v", err)
				return
			}
			// deploy remote file to local.
			if err = os.WriteFile(fPath, content, os.ModePerm); err != nil {
				gprint.PrintError("Write file failed: %+v", err)
				return
			}
		}
	} else {
		gprint.PrintError("download failed: %s", remoteFileName)
	}
}
