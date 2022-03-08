package ansible

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/KubeOperator/kobe/api"
	"github.com/KubeOperator/kobe/pkg/constant"
	"github.com/KubeOperator/kobe/pkg/util"
	"github.com/prometheus/common/log"
)

type PlaybookRunner struct {
	Project  *api.Project
	Playbook string
	Tag      string
}

type AdhocRunner struct {
	Module  string
	Param   string
	Pattern string
}

func (a *AdhocRunner) Run(ch chan []byte, result *api.KobeResult) {
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
	if util.CheckIllegal(ansiblePath, inventoryProviderPath, a.Pattern, a.Module) {
		result.Success = false
		result.Message = "license contains invalid characters!"
		return
	}
	cmd := exec.Command(ansiblePath, "-e", "host_key_checking=False", "-i", inventoryProviderPath, a.Pattern, "-m", a.Module)
	if a.Param != "" {
		cmd.Args = append(cmd.Args, "-a", a.Param)
	}
	cmdEnv := make([]string, 0)
	cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", constant.TaskEnvKey, result.Id))
	cmd.Env = append(os.Environ(), cmdEnv...)
	log.Infof("id:%s  content :%s", result.Id, cmd.String())
	runCmd(ch, "adhoc", cmd, result)

}

func (p *PlaybookRunner) Run(ch chan []byte, result *api.KobeResult) {
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

	itemPath := path.Join(constant.ProjectDir, p.Project.Name, p.Playbook)
	if util.CheckIllegal(ansiblePath, inventoryProviderPath, itemPath) {
		result.Success = false
		result.Message = "license contains invalid characters!"
		return
	}
	cmd := exec.Command(ansiblePath, "-i", inventoryProviderPath, itemPath)
	varPath := path.Join(constant.ProjectDir, p.Project.Name, constant.AnsibleVariablesName)
	exists, _ := util.PathExists(varPath)
	if exists {
		varPath = "@" + varPath
		cmd.Args = append(cmd.Args, "-e", varPath)
	}
	if p.Tag != "" {
		cmd.Args = append(cmd.Args, "-t", p.Tag)
	}
	cmdEnv := make([]string, 0)
	cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", constant.TaskEnvKey, result.Id))
	cmd.Env = append(os.Environ(), cmdEnv...)
	log.Infof("id:%s  content :%s", result.Id, cmd.String())
	runCmd(ch, p.Project.Name, cmd, result)
}

func runCmd(ch chan []byte, projectName string, cmd *exec.Cmd, result *api.KobeResult) {
	if ch == nil {
		return
	}
	workPath, err := initWorkSpace(projectName)
	if err != nil {
		result.Message = err.Error()
		return
	}
	pwd, err := os.Getwd()
	if err != nil {
		result.Message = err.Error()
		return
	}
	if err := os.Chdir(workPath); err != nil {
		result.Message = err.Error()
		return
	}
	defer func() {
		if err := os.Chdir(pwd); err != nil {
			result.Message = err.Error()
		}
		result.EndTime = time.Now().String()
	}()
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

func initWorkSpace(projectName string) (string, error) {
	workPath := path.Join(constant.WorkDir, projectName)
	if err := os.MkdirAll(workPath, 0750); err != nil {
		return "", err
	}
	return workPath, nil
}
