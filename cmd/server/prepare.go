package main

import (
	"github.com/KubeOperator/kobe/pkg/constant"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"text/template"
)

func prepareStart() error {
	funcs := []func() error{
		makeDataDir,
		makeCacheDir,
		makeKeyDir,
		lookUpAnsibleBinPath,
		lookUpKobeInventoryBinPath,
		cleanWorkPath,
		renderConfig,
	}
	for _, f := range funcs {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func makeDataDir() error {
	return os.MkdirAll(constant.ProjectDir, 0755)

}

func makeCacheDir() error {
	return os.MkdirAll(constant.CacheDir, 0755)
}

func makeKeyDir() error {
	return os.MkdirAll(constant.KeyDir, 0755)
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
	_ = os.RemoveAll(constant.WorkDir)
	return nil
}

func renderConfig() error {
	tmpl := constant.AnsibleTemplateFilePath
	file, err := os.OpenFile(constant.AnsibleConfPath, os.O_CREATE|os.O_RDWR, 0755)
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
