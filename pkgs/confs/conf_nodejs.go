package confs

import (
	"os"

	tui "github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/gvc/pkgs/utils"
)

type NodejsConf struct {
	CompilerUrl string   `koanf:"compiler_url"`
	ReleaseUrl  string   `koanf:"release_url"`
	ProxyUrls   []string `koanf:"proxy_urls"`
	path        string
}

func NewNodejsConf() (r *NodejsConf) {
	r = &NodejsConf{
		path: NodejsFilesDir,
	}
	r.setup()
	return
}

func (that *NodejsConf) setup() {
	if ok, _ := utils.PathIsExist(that.path); !ok {
		if err := os.MkdirAll(that.path, os.ModePerm); err != nil {
			tui.PrintError(err)
		}
	}
}

func (that *NodejsConf) Reset() {
	that.CompilerUrl = "https://nodejs.org/dist/index.json"
	that.ReleaseUrl = "https://nodejs.org/download/release"
	that.ProxyUrls = []string{
		"https://registry.npm.taobao.org",
		"https://registry.npmmirror.com/",
		"https://mirrors.huaweicloud.com/repository/npm/",
		"https://registry.npmjs.org/",
	}
}
