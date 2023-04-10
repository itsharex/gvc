package vctrl

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gookit/color"
	"github.com/mholt/archiver/v3"
	config "github.com/moqsien/gvc/pkgs/confs"
	"github.com/moqsien/gvc/pkgs/downloader"
	"github.com/moqsien/gvc/pkgs/utils"
	"github.com/moqsien/gvc/pkgs/utils/sorts"
)

type JuliaPackage struct {
	Url      string
	FileName string
	OS       string
	Arch     string
	Checksum string
}

type JuliaVersion struct {
	Versions map[string][]*JuliaPackage
	Json     *gjson.Json
	Conf     *config.GVConfig
	c        *colly.Collector
	d        *downloader.Downloader
	env      *utils.EnvsHandler
}

func NewJuliaVersion() (jv *JuliaVersion) {
	jv = &JuliaVersion{
		Versions: make(map[string][]*JuliaPackage, 500),
		Conf:     config.New(),
		c:        colly.NewCollector(),
		d:        &downloader.Downloader{},
		env:      utils.NewEnvsHandler(),
	}
	jv.initeDirs()
	return
}

func (that *JuliaVersion) initeDirs() {
	if ok, _ := utils.PathIsExist(config.JuliaRootDir); !ok {
		os.RemoveAll(config.JuliaRootDir)
		if err := os.MkdirAll(config.JuliaRootDir, os.ModePerm); err != nil {
			fmt.Println("[mkdir Failed] ", err)
		}
	}
	if ok, _ := utils.PathIsExist(config.JuliaTarFilePath); !ok {
		if err := os.MkdirAll(config.JuliaTarFilePath, os.ModePerm); err != nil {
			fmt.Println("[mkdir Failed] ", err)
		}
	}
	if ok, _ := utils.PathIsExist(config.JuliaUntarFilePath); !ok {
		if err := os.MkdirAll(config.JuliaUntarFilePath, os.ModePerm); err != nil {
			fmt.Println("[mkdir Failed] ", err)
		}
	}
}

func (that *JuliaVersion) getJson() {
	jUrl := that.Conf.Julia.VersionUrl
	if !utils.VerifyUrls(jUrl) {
		return
	}
	that.c.OnResponse(func(r *colly.Response) {
		that.Json = gjson.New(r.Body)
	})
	that.c.Visit(jUrl)
}

func (that *JuliaVersion) GetVersions() {
	if that.Json == nil {
		that.getJson()
	}
	if that.Json != nil {
		m := that.Json.GetMap(".")
		for version, vcontent := range m {
			j := gjson.New(vcontent)
			if j.GetBool("stable") {
				if len(that.Versions[version]) == 0 {
					that.Versions[version] = []*JuliaPackage{}
				}
				for _, f := range j.GetArray("files") {
					fj := gjson.New(f)
					if fj.GetString("kind") == "archive" {
						fext := fj.GetString("extension")
						if fext != "tar.gz" && fext != "zip" && fext != "tar.xz" {
							continue
						}
						p := &JuliaPackage{}
						p.Url = fj.GetString("url")
						p.Arch = utils.ParseArch(fj.GetString("arch"))
						p.OS = utils.ParsePlatform(fj.GetString("os"))
						if p.Arch == "" || p.OS == "" || p.Url == "" {
							continue
						}
						p.Checksum = fj.GetString("sha256")

						p.FileName = fmt.Sprintf("julia-%s-%s-%s.%s",
							version, p.OS, p.Arch, fext)
						that.Versions[version] = append(that.Versions[version], p)
					}
				}
			}
		}
	}
}

func (that *JuliaVersion) ShowVersions() {
	if len(that.Versions) == 0 {
		that.GetVersions()
	}
	vList := []string{}
	for v := range that.Versions {
		vList = append(vList, v)
	}
	res := sorts.SortGoVersion(vList)
	fmt.Println(strings.Join(res, "  "))
}

func (that *JuliaVersion) findPackage(version string) *JuliaPackage {
	for _, pk := range that.Versions[version] {
		if pk.Arch == runtime.GOARCH && pk.OS == runtime.GOOS {
			if pk.Url != "" && that.Conf.Julia.BaseUrl != "" {
				uList := strings.Split(pk.Url, "bin/")
				if len(uList) > 1 {
					pk.Url, _ = url.JoinPath(that.Conf.Julia.BaseUrl, uList[1])
				}
			}
			return pk
		}
	}
	return nil
}

func (that *JuliaVersion) download(version string) (r string) {
	if len(that.Versions) == 0 {
		that.GetVersions()
	}

	if p := that.findPackage(version); p != nil {
		that.d.Url = p.Url
		if !utils.VerifyUrls(that.d.Url) {
			return
		}
		that.d.Timeout = 100 * time.Minute
		fpath := filepath.Join(config.JuliaTarFilePath, p.FileName)
		if size := that.d.GetFile(fpath, os.O_CREATE|os.O_WRONLY, 0644); size > 0 {
			if p.Checksum != "" {
				if ok := utils.CheckFile(fpath, "sha256", p.Checksum); ok {
					return fpath
				} else {
					os.RemoveAll(fpath)
				}
			} else {
				return fpath
			}
		} else {
			os.RemoveAll(fpath)
		}
	} else {
		fmt.Println("Invalid Julia version. ", version)
	}
	return
}

func (that *JuliaVersion) CheckAndInitEnv() {
	if runtime.GOOS != utils.Windows {
		juliaEnv := fmt.Sprintf(utils.JuliaEnv,
			config.JuliaRootDir,
			that.Conf.Julia.PkgServer)
		that.env.UpdateSub(utils.SUB_JULIA, juliaEnv)
	} else {
		envList := map[string]string{
			"JULIA_PKG_SERVER": that.Conf.Julia.PkgServer,
			"PATH":             filepath.Join(config.JuliaRootDir, "bin"),
		}
		that.env.SetEnvForWin(envList)
	}
}

func (that *JuliaVersion) UseVersion(version string) {
	untarfile := filepath.Join(config.JuliaUntarFilePath, version)
	if ok, _ := utils.PathIsExist(untarfile); !ok {
		if tarfile := that.download(version); tarfile != "" {
			if err := archiver.Unarchive(tarfile, untarfile); err != nil {
				os.RemoveAll(untarfile)
				fmt.Println("[Unarchive failed] ", err)
				return
			}
		}
	}
	if ok, _ := utils.PathIsExist(config.JuliaRootDir); ok {
		os.RemoveAll(config.JuliaRootDir)
	}
	finder := utils.NewBinaryFinder(untarfile, "bin")
	dir := finder.String()
	if dir != "" {
		if err := utils.MkSymLink(dir, config.JuliaRootDir); err != nil {
			fmt.Println("[Create link failed] ", err)
			return
		}
		if !that.env.DoesEnvExist(utils.SUB_JULIA) {
			that.CheckAndInitEnv()
		}
		utils.RecordVersion(version, dir)
		fmt.Println("Use", version, "succeeded!")
	}
}

func (that *JuliaVersion) ShowInstalled() {
	current := utils.ReadVersion(config.JuliaRootDir)
	dList, _ := os.ReadDir(config.JuliaUntarFilePath)
	for _, d := range dList {
		if d.IsDir() {
			switch d.Name() {
			case current:
				s := fmt.Sprintf("%s <Current>", d.Name())
				color.Yellow.Println(s)
			default:
				color.Cyan.Println(d.Name())
			}
		}
	}
}

func (that *JuliaVersion) removeTarFile(version string) {
	fName := fmt.Sprintf("julia-%s-%s-%s", version, runtime.GOOS, runtime.GOARCH)
	dList, _ := os.ReadDir(config.JuliaTarFilePath)
	for _, d := range dList {
		if !d.IsDir() && strings.Contains(d.Name(), fName) {
			os.RemoveAll(filepath.Join(config.JuliaTarFilePath, d.Name()))
		}
	}
}

func (that *JuliaVersion) RemoveVersion(version string) {
	current := utils.ReadVersion(config.JuliaRootDir)
	if version == current {
		return
	}
	dList, _ := os.ReadDir(config.JuliaUntarFilePath)
	for _, d := range dList {
		if d.IsDir() && d.Name() == version {
			os.RemoveAll(filepath.Join(config.JuliaUntarFilePath, d.Name()))
			that.removeTarFile(version)
		}
	}
}

func (that *JuliaVersion) RemoveUnused() {
	current := utils.ReadVersion(config.JuliaRootDir)
	dList, _ := os.ReadDir(config.JuliaUntarFilePath)
	for _, d := range dList {
		if d.IsDir() && d.Name() != current {
			os.RemoveAll(filepath.Join(config.JuliaUntarFilePath, d.Name()))
			that.removeTarFile(d.Name())
		}
	}
}