package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/KubeOperator/kobe/api"
	"github.com/KubeOperator/kobe/pkg/constant"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
)

type Result map[string]map[string]interface{}

type kobeInventoryProvider struct {
	host string
	port int
}

func NewKobeInventoryProvider(host string, port int) *kobeInventoryProvider {
	return &kobeInventoryProvider{
		host: host,
		port: port,
	}
}
func (r Result) String() string {
	b, err := json.Marshal(&r)
	if err != nil {
		return ""
	}
	return string(b)
}

func (kip kobeInventoryProvider) getInventory(id string) (*api.Inventory, error) {

	conn, err := kip.createConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := api.NewKobeApiClient(conn)
	request := &api.GetInventoryRequest{
		Id: id,
	}
	resp, err := client.GetInventory(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return resp.Item, nil
}

func (kip kobeInventoryProvider) ListHandler() (Result, error) {
	id, err := kip.getInventoryId()
	if err != nil {
		return nil, err
	}
	inventory, _ := kip.getInventory(id)

	if inventory == nil {
		return nil, fmt.Errorf("can not find inventory in cache invalid taskId %s", id)
	}

	groups := make(map[string]map[string]interface{})
	all := api.Group{
		Name:     "all",
		Hosts:    []string{},
		Children: []string{},
		Vars:     inventory.Vars,
	}
	for _, group := range inventory.Groups {
		m := parseGroupToMap(group)
		groups[group.Name] = m
		all.Children = append(all.Children, group.Name)
	}
	meta := map[string]interface{}{}
	hostVars := map[string]interface{}{}
	for _, host := range inventory.Hosts {
		all.Hosts = append(all.Hosts, host.Name)
		hostVars[host.Name] = map[string]interface{}{
			"ansible_ssh_host": host.Ip,
			"ansible_ssh_port": host.Port,
			"ansible_ssh_user": host.User,
		}
		if host.Password != "" {
			m := hostVars[host.Name].(map[string]interface{})
			m["ansible_ssh_pass"] = host.Password
		}
		if host.PrivateKey != "" {
			m := hostVars[host.Name].(map[string]interface{})
			m["ansible_ssh_private_key_file"] = generatePrivateKeyFile(host.Name, host.PrivateKey)
		}
		if host.ProxyConfig != nil && host.ProxyConfig.Enable {
			m := hostVars[host.Name].(map[string]interface{})
			m["ansible_ssh_common_args"] = fmt.Sprintf("-o ProxyCommand=\"sshpass -p %s ssh -W  %%h:%%p -p %d -q %s@%s\" -o StrictHostKeyChecking=no", host.ProxyConfig.Password, host.ProxyConfig.Port, host.ProxyConfig.User, host.ProxyConfig.Ip)
		}
		if host.Vars != nil {
			m := hostVars[host.Name].(map[string]interface{})
			for k, v := range host.Vars {
				m[k] = v
			}
		}
	}
	groups[all.Name] = parseGroupToMap(&all)
	meta["hostvars"] = hostVars
	groups["_meta"] = meta
	return groups, nil
}

func parseGroupToMap(group *api.Group) map[string]interface{} {
	m := map[string]interface{}{}
	if group.Hosts != nil {
		m["hosts"] = group.Hosts
	} else {
		m["hosts"] = []string{}
	}
	if group.Children != nil {
		m["children"] = group.Children
	} else {
		m["children"] = []string{}
	}
	if group.Vars != nil {
		m["vars"] = group.Vars
	} else {
		m["vars"] = map[string]string{}
	}
	return m
}

func (kip kobeInventoryProvider) getInventoryId() (string, error) {
	id := os.Getenv(constant.TaskEnvKey)
	if id == "" {
		return "", fmt.Errorf("invalid id: %s", id)
	}
	return id, nil
}

func (kip kobeInventoryProvider) createConnection() (*grpc.ClientConn, error) {
	address := fmt.Sprintf("%s:%d", kip.host, kip.port)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func generatePrivateKeyFile(hostName string, content string) string {
	fileName := fmt.Sprintf("%s-%s.pem", hostName, uuid.NewV4().String())
	p := path.Join(constant.KeyDir, fileName)
	f, err := os.OpenFile(p, os.O_CREATE, 0600)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer f.Close()
	err = ioutil.WriteFile(f.Name(), []byte(content), 0600)
	if err != nil {
		log.Println(err)
		return ""
	}
	return p
}
