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
	"sync"
)

type PlaybookRunner struct {
	Project  api.Project
	Playbook string
	Tag      string
}

type AdhocRunner struct {
	Module  string
	Param   string
	Pattern string
}

func (a *AdhocRunner) Run(wg *sync.WaitGroup, result *api.Result) io.ReadCloser {
	ansiblePath, err := exec.LookPath(constant.AnsibleBinPath)
	if err != nil {
		result.Success = false
		result.Message = err.Error()
		return nil
	}
	inventoryProviderPath, err := exec.LookPath(constant.InventoryProviderBinPath)
	if err != nil {
		result.Success = false
		result.Message = err.Error()
		return nil
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
	return runCmd(wg, cmd, result)

}

func (p *PlaybookRunner) Run(wg *sync.WaitGroup, result *api.Result) io.ReadCloser {
	ansiblePath, err := exec.LookPath(constant.AnsiblePlaybookBinPath)
	if err != nil {
		result.Success = false
		result.Message = err.Error()
		return nil
	}
	inventoryProviderPath, err := exec.LookPath(constant.InventoryProviderBinPath)
	if err != nil {
		result.Success = false
		result.Message = err.Error()
		return nil
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
	if p.Tag != "" {
		cmd.Args = append(cmd.Args, "-t", p.Tag)
	}
	cmdEnv := make([]string, 0)
	cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", constant.TaskEnvKey, result.Id))
	cmd.Env = append(os.Environ(), cmdEnv...)
	log.Infof("id:%s  content :%s", result.Id, cmd.String())
	return runCmd(wg, cmd, result)
}

func runCmd(wg *sync.WaitGroup, cmd *exec.Cmd, result *api.Result) io.ReadCloser {
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		result.Success = false
		result.Message = err.Error()
		wg.Done()
		return nil
	}
	if err := cmd.Start(); err != nil {
		result.Success = false
		result.Message = err.Error()
		wg.Done()
		return nil
	}
	go func() {
		if err = cmd.Wait(); err != nil {
			b, err := ioutil.ReadAll(stderr)
			if err != nil {
				log.Error(err)
			}
			result.Success = false
			result.Message = string(b)
			wg.Done()
			return
		}
		result.Success = true
		wg.Done()
	}()
	return stdout
}
