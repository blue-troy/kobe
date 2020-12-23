package constant

import (
	"github.com/spf13/viper"
	"path"
)

const (
	InventoryProviderBinPath = "kobe-inventory"
	AnsiblePlaybookBinPath   = "ansible-playbook"
	AnsibleBinPath           = "ansible"
	TaskEnvKey               = "KO_TASK_ID"
	AnsibleVariablesName     = "variables.yml"
)

var (
	BaseDir                 = "/var/kobe"
	DataDir                 = path.Join(BaseDir, "data")
	CacheDir                = path.Join(DataDir, "cache")
	KeyDir                  = path.Join(DataDir, "key")
	WorkDir                 = path.Join(BaseDir, "work")
	ProjectDir              = path.Join(DataDir, "project")
	AnsibleTemplateFilePath = path.Join("/", "etc", "kobe", "ansible.cfg.tmpl")
	AnsibleConfPath         = path.Join("/", "etc", "ansible", "ansible.cfg")
)

func Init() {
	BaseDir = viper.GetString("base")
}
