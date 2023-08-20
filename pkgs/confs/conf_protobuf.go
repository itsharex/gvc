package confs

import (
	"os"

	tui "github.com/moqsien/goutils/pkgs/gtui"
	"github.com/moqsien/gvc/pkgs/utils"
)

type ProtobufConf struct {
	GitlabUrls map[string]string `koanf:"gitlab_urls"`
	path       string
}

func NewProtobuf() (r *ProtobufConf) {
	r = &ProtobufConf{
		path: ProtobufDir, // TODO: modification
	}
	r.setup()
	return
}

func (that *ProtobufConf) setup() {
	if ok, _ := utils.PathIsExist(that.path); !ok {
		if err := os.MkdirAll(that.path, os.ModePerm); err != nil {
			tui.PrintError(err)
		}
	}
}

func (that *ProtobufConf) Reset() {
	that.GitlabUrls = map[string]string{
		utils.Windows: "https://gitlab.com/moqsien/gvc_resources/-/raw/main/protoc_win64.zip",
		"linux_amd64": "https://gitlab.com/moqsien/gvc_resources/-/raw/main/protoc_linux_x86_64.zip",
		"linux_arm64": "https://gitlab.com/moqsien/gvc_resources/-/raw/main/protoc_linux_aarch_64.zip",
		utils.MacOS:   "https://gitlab.com/moqsien/gvc_resources/-/raw/main/protoc_osx_universal_binary.zip",
	}
}
