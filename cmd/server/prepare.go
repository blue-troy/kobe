package main

import (
	"os"
	"os/exec"
	"text/template"

	"github.com/KubeOperator/kobe/pkg/constant"
	"github.com/spf13/viper"
)

func prepareStart() error {
	funcs := []func() error{
		makeDataDir,
		makeCacheDir,
		makeKeyDir,
		makeAnsibleCfgDir,
		lookUpAnsibleBinPath,
		lookUpKobeInventoryBinPath,
		cleanWorkPath,
		randerAnsibleConfig,
	}
	for _, f := range funcs {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func makeDataDir() error {
	return os.MkdirAll(constant.ProjectDir, 0750)

}

func makeAnsibleCfgDir() error {
	return os.MkdirAll(constant.AnsibleConfDir, 0750)
}

func makeCacheDir() error {
	return os.MkdirAll(constant.CacheDir, 0750)
}

func makeKeyDir() error {
	return os.MkdirAll(constant.KeyDir, 0750)
}

func lookUpAnsibleBinPath() error {
	_, err := exec.LookPath(constant.AnsiblePlaybookBinPath)
	if err != nil {
		return err
	}
	return nil
}

func lookUpKobeInventoryBinPath() error {
	_, err := exec.LookPath(constant.InventoryProviderBinPath)
	if err != nil {
		return err
	}
	return nil
}

func cleanWorkPath() error {
	if err := os.RemoveAll(constant.WorkDir); err != nil {
		return err
	}
	return nil
}

func randerAnsibleConfig() error {
	tmpl := constant.AnsibleTemplateFilePath
	file, err := os.OpenFile(constant.AnsibleConfPath, os.O_CREATE|os.O_RDWR, 0750)
	if err != nil {
		return err
	}
	defer file.Close()
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		return err
	}
	data := viper.GetStringMap("ansible")
	if err := t.Execute(file, data); err != nil {
		return err
	}
	return nil
}
