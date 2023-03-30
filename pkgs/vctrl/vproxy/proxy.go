package vproxy

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	config "github.com/moqsien/gvc/pkgs/confs"
	"github.com/moqsien/gvc/pkgs/utils"
	"github.com/xtls/xray-core/core"
	xconf "github.com/xtls/xray-core/infra/conf"
	"github.com/xtls/xray-core/infra/conf/serial"
	_ "github.com/xtls/xray-core/main/confloader/external"
	_ "github.com/xtls/xray-core/main/distro/all"
)

type ProxyType string

const (
	Vmess  ProxyType = "vmess"
	SS     ProxyType = "ss"
	Trojan ProxyType = "trojan"
)

type Proxy struct {
	Uri string `koanf:"uri"`
	Rtt int    `koanf:"rtt"`
}

type ProxyList struct {
	Proxies []*Proxy `koanf:"proxies"`
	Type    ProxyType
	Date    string `koanf:"date"`
	k       *koanf.Koanf
	parser  *json.JSON
	path    string
}

func (that *ProxyList) GetDate() string {
	return time.Now().Format("2006-01-02")
}

func (that *ProxyList) Reload() {
	if ok, _ := utils.PathIsExist(that.path); !ok {
		fmt.Println("ProxyList file does not exist.")
		return
	}
	err := that.k.Load(file.Provider(that.path), that.parser)
	if err != nil {
		fmt.Println("[Config Load Failed] ", err)
		return
	}
	that.k.UnmarshalWithConf("", that, koanf.UnmarshalConf{Tag: "koanf"})
}

func (that *ProxyList) restore() {
	if ok, _ := utils.PathIsExist(config.ProxyFilesDir); !ok {
		os.MkdirAll(config.ProxyFilesDir, os.ModePerm)
	}
	that.k.Load(structs.Provider(*that, "koanf"), nil)
	if b, err := that.k.Marshal(that.parser); err == nil && len(b) > 0 {
		os.WriteFile(that.path, b, 0666)
	}
}

func (that *ProxyList) Update(proxies []*Proxy) {
	if len(proxies) == 0 {
		fmt.Println("[Proxy List is empty]")
		return
	}
	that.Proxies = proxies
	that.Date = that.GetDate()
	that.restore()
}

type Proxyer struct {
	Conf       *config.GVConfig
	XRay       *core.Instance
	XRayConfig *xconf.Config
	ProxyList  *ProxyList
	c          *colly.Collector
	filter     map[string]struct{}
}

func NewProxyer() (r *Proxyer) {
	r = &Proxyer{
		Conf: config.New(),
		c:    colly.NewCollector(),
		ProxyList: &ProxyList{
			Proxies: make([]*Proxy, 200),
			k:       koanf.New("."),
			parser:  json.Parser(),
			path:    config.ProxyListFilePath,
		},
		XRayConfig: &xconf.Config{},
	}
	return
}

func (that *Proxyer) decodeUri(uri string) (r string) {
	if strings.HasPrefix(uri, "vmess://") {
		sList := strings.Split(uri, "://")
		uri = sList[1]
		s, _ := base64.StdEncoding.DecodeString(uri)
		r = string(s)
		r = fmt.Sprintf("vmess|||%s", r)
	} else {
		r = uri
	}
	return
}

func (that *Proxyer) decodeStr(rawStr string) (res string) {
	s, _ := base64.StdEncoding.DecodeString(rawStr)
	res = string(s)
	return
}

func (that *Proxyer) parseProxy(body []byte, pType string) (result []*Proxy) {
	r := string(body)
	if !strings.Contains(r, "vmess") {
		r = that.decodeStr(r)
	}
	if strings.Contains(r, "vmess") {
		for _, p := range strings.Split(r, "\n") {
			pUrl := strings.Trim(p, " ")
			if !strings.HasPrefix(pUrl, "vmess") {
				fmt.Println(pUrl)
				continue
			}
			if _, ok := that.filter[pUrl]; !ok {
				that.filter[pUrl] = struct{}{}
				result = append(result, &Proxy{
					Uri: that.decodeUri(pUrl),
				})
			}
		}
	}
	return
}

func (that *Proxyer) GetProxyList(force bool) {
	that.ProxyList.Reload()
	if that.ProxyList.GetDate() != that.ProxyList.Date || force {
		that.filter = map[string]struct{}{}
		pList := []*Proxy{}
		for _, url := range that.Conf.Proxy.GetSubUrls() {
			// that.collector.SetRequestTimeout(5 * time.Second)
			that.c.OnResponse(func(r *colly.Response) {
				res := that.parseProxy(r.Body, string(Vmess))
				pList = append(pList, res...)
			})
			that.c.Visit(url)
		}
		that.ProxyList.Update(pList)
		that.ProxyList.Type = Vmess
	}
	that.ProxyList.Reload()
}

func (that *Proxyer) LoadXrayConfig() (err error) {
	r := utils.ConvertStrToReader(XrayVmessConfStr)
	c := &xconf.Config{}
	if err == nil {
		c, err = serial.DecodeJSONConfig(r)
	}
	that.XRayConfig = c
	return err
}

func (that *Proxyer) RunXray() {
	err := that.LoadXrayConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	var c *core.Config
	c, err = that.XRayConfig.Build()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(c.Outbound))
	server, err := core.New(c)
	if err != nil {
		fmt.Println("2222", err)
		return
	}
	fmt.Println("=====")
	server.Start()
	defer server.Close()
}