package ansible

import (
	"bytes"
	"fmt"
	"github.com/KubeOperator/kobe/api"
	"github.com/KubeOperator/kobe/pkg/constant"
	"github.com/KubeOperator/kobe/pkg/util"
	"github.com/prometheus/common/log"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

type PlaybookRunner struct {
	Project  api.Project
	Playbook string
}

type AdhocRunner struct {
	Module  string
	Param   string
	Pattern string
}

func (a *AdhocRunner) Run(ch chan []byte, result *api.Result) {
	ansiblePath, err := exec.LookPath(constant.AnsibleBinPath)
	if err != nil {
		result.Success = false
		result.Message = err.Error()
		return
	}
	inventoryProviderPath, err := exec.LookPath(constant.InventoryProviderBinPath)
	if err != nil {
		result.Success = false
		result.Message = err.Error()
		return
	}
	cmd := exec.Command(ansiblePath,
		"-e", "host_key_checking=False",
		"-i", inventoryProviderPath, a.Pattern, "-m", a.Module)
	if a.Param != "" {
		cmd.Args = append(cmd.Args, "-a", a.Param)
	}
	cmdEnv := make([]string, 0)
	cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", constant.TaskEnvKey, result.Id))
	cmd.Env = append(os.Environ(), cmdEnv...)
	log.Infof("id:%s  content :%s", result.Id, cmd.String())
	runCmd(ch, cmd, result)

}

func (p *PlaybookRunner) Run(ch chan []byte, result *api.Result) {
	ansiblePath, err := exec.LookPath(constant.AnsiblePlaybookBinPath)
	if err != nil {
		result.Success = false
		result.Message = err.Error()
		return
	}
	inventoryProviderPath, err := exec.LookPath(constant.InventoryProviderBinPath)
	if err != nil {
		result.Success = false
		result.Message = err.Error()
		return
	}

	cmd := exec.Command(ansiblePath,
		"-i", inventoryProviderPath,
		path.Join(constant.ProjectDir, p.Project.Name, p.Playbook))
	varPath := path.Join(constant.ProjectDir, p.Project.Name, constant.AnsibleVariablesName)
	exists, _ := util.PathExists(varPath)
	if exists {
		varPath = "@" + varPath
		cmd.Args = append(cmd.Args, "-e", varPath)
	}
	cmdEnv := make([]string, 0)
	cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", constant.TaskEnvKey, result.Id))
	cmd.Env = append(os.Environ(), cmdEnv...)
	log.Infof("id:%s  content :%s", result.Id, cmd.String())
	runCmd(ch, cmd, result)
}

func runCmd(ch chan []byte, cmd *exec.Cmd, result *api.Result) {
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		result.Success = false
		result.Message = err.Error()
		return
	}
	if err := cmd.Start(); err != nil {
		result.Success = false
		result.Message = err.Error()
		return
	}
	buf := make([]byte, 4096)
	for {
		nr, err := stdout.Read(buf)
		if nr > 0 {
			select {
			case ch <- buf[:nr]:
			default:
			}
		}
		if err != nil || io.EOF == err {
			break
		}
	}
	close(ch)
	if err = cmd.Wait(); err != nil {
		result.Success = false
		b, err := ioutil.ReadAll(stderr)
		if err != nil {
			log.Error(err)
			return
		}
		result.Message = string(b)
		return
	}
	result.Success = true
}
