package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/KubeOperator/kobe/api"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/common/log"
	uuid "github.com/satori/go.uuid"
	"io"
	"sync"
	"time"
)

type Kobe struct {
	taskCache      *cache.Cache
	inventoryCache *cache.Cache
	pool           *Pool
	resultCache    *cache.Cache
	stdoutCache    *cache.Cache
}

func NewKobe() *Kobe {
	return &Kobe{
		taskCache:      cache.New(24*time.Hour, 5*time.Minute),
		inventoryCache: cache.New(10*time.Minute, 5*time.Minute),
		resultCache:    cache.New(24*time.Hour, 5*time.Minute),
		stdoutCache:    cache.New(24*time.Hour, 5*time.Minute),
		pool:           NewPool(),
	}
}

func (k *Kobe) SetAnsibleResult(context context.Context, req *api.SetAnsibleResultRequest) (*api.SetAnsibleResultResponse, error) {
	k.resultCache.Set(req.TaskId, req.Result, cache.DefaultExpiration)
	log.Infof("receive a result %s", req.TaskId)
	return &api.SetAnsibleResultResponse{}, nil
}

func (k *Kobe) TestHello(ctx context.Context, req *api.HelloRequest) (*api.HelloResponse, error) {
	println("rec  message %s", req.Name)
	return &api.HelloResponse{Name: "world"}, nil
}

func (k *Kobe) CreateProject(ctx context.Context, req *api.CreateProjectRequest) (*api.CreateProjectResponse, error) {
	pm := ProjectManager{}
	p, err := pm.CreateProject(req.Name, req.Source)
	if err != nil {
		return nil, err
	}
	resp := &api.CreateProjectResponse{
		Item: p,
	}
	return resp, nil
}
func (k *Kobe) ListProject(ctx context.Context, req *api.ListProjectRequest) (*api.ListProjectResponse, error) {
	pm := ProjectManager{}
	ps, err := pm.SearchProjects()
	if err != nil {
		return nil, err
	}
	resp := &api.ListProjectResponse{
		Items: ps,
	}
	return resp, nil
}

func (k *Kobe) GetInventory(ctx context.Context, req *api.GetInventoryRequest) (*api.GetInventoryResponse, error) {
	item, _ := k.inventoryCache.Get(req.Id)
	if item == nil {
		return nil, errors.New("inventory is expire")
	}
	resp := &api.GetInventoryResponse{
		Item: item.(*api.Inventory),
	}
	return resp, nil
}

func (k *Kobe) WatchResult(req *api.WatchRequest, server api.KobeApi_WatchResultServer) error {
	stdout, found := k.stdoutCache.Get(req.TaskId)
	if !found {
		return errors.New(fmt.Sprintf("can not find task: %s", req.TaskId))
	}
	val, ok := stdout.(io.ReadCloser)
	if !ok {
		return errors.New(fmt.Sprintf("invalid cache"))
	}
	buf := make([]byte, 4096)
	for {
		nr, err := val.Read(buf)
		if nr > 0 {
			_ = server.Send(&api.WatchStream{
				Stream: buf[:nr],
			})
		}
		if err != nil || io.EOF == err {
			break
		}
	}
	return nil
}

func (k *Kobe) RunAdhoc(ctx context.Context, req *api.RunAdhocRequest) (*api.RunAdhocResult, error) {
	rm := RunnerManager{
		inventoryCache: k.inventoryCache,
	}
	id := uuid.NewV4().String()
	result := api.Result{
		Id:        id,
		StartTime: time.Now().Format("2006-01-02 15:04:05"),
		EndTime:   "",
		Message:   "",
		Success:   false,
		Finished:  false,
		Running:   false,
		Content:   "",
	}
	k.taskCache.Set(result.Id, &result, cache.DefaultExpiration)
	k.inventoryCache.Set(result.Id, req.Inventory, cache.DefaultExpiration)
	runner, err := rm.CreateAdhocRunner(req.Pattern, req.Module, req.Param)
	if err != nil {
		return nil, err
	}
	k.taskCache.Set(result.Id, &result, cache.DefaultExpiration)
	task := func() {
		wg := sync.WaitGroup{}
		wg.Add(1)
		stdout := runner.Run(&wg, &result)
		result.Running = true
		k.taskCache.Set(result.Id, &result, cache.DefaultExpiration)
		k.stdoutCache.Set(result.Id, stdout, cache.DefaultExpiration)
		wg.Wait()
		result.Finished = true
		result.EndTime = time.Now().Format("2006-01-02 15:04:05")
		k.taskCache.Set(result.Id, &result, cache.DefaultExpiration)
	}
	k.pool.Commit(task)
	return &api.RunAdhocResult{
		Result: &result,
	}, nil
}

func (k *Kobe) RunPlaybook(ctx context.Context, req *api.RunPlaybookRequest) (*api.RunPlaybookResult, error) {
	rm := RunnerManager{
		inventoryCache: k.inventoryCache,
	}
	id := uuid.NewV4().String()
	result := api.Result{
		Id:        id,
		StartTime: time.Now().Format("2006-01-02 15:04:05"),
		EndTime:   "",
		Message:   "",
		Success:   false,
		Finished:  false,
		Content:   "",
		Project:   req.Project,
		Running:   false,
	}
	k.taskCache.Set(result.Id, &result, cache.DefaultExpiration)
	k.inventoryCache.Set(result.Id, req.Inventory, cache.DefaultExpiration)
	runner, err := rm.CreatePlaybookRunner(req.Project, req.Playbook, req.Tag)
	if err != nil {
		return nil, err
	}
	k.taskCache.Set(result.Id, &result, cache.DefaultExpiration)
	task := func() {
		wg := sync.WaitGroup{}
		wg.Add(1)
		stdout := runner.Run(&wg, &result)
		result.Running = true
		k.taskCache.Set(result.Id, &result, cache.DefaultExpiration)
		k.stdoutCache.Set(result.Id, stdout, cache.DefaultExpiration)
		wg.Wait()
		result.Finished = true
		result.EndTime = time.Now().Format("2006-01-02 15:04:05")
		k.taskCache.Set(result.Id, &result, cache.DefaultExpiration)
	}
	k.pool.Commit(task)
	return &api.RunPlaybookResult{
		Result: &result,
	}, nil
}

func (k *Kobe) GetResult(ctx context.Context, req *api.GetResultRequest) (*api.GetResultResponse, error) {
	id := req.GetTaskId()
	result, found := k.taskCache.Get(id)
	if !found {
		return nil, errors.New(fmt.Sprintf("can not find task: %s result", id))
	}
	val, ok := result.(*api.Result)
	if !ok {
		return nil, errors.New("invalid result type")
	}
	if val.Finished {
		content, found := k.resultCache.Get(id)
		if !found {
			return nil, fmt.Errorf("can not found %s result in cache", id)
		}
		resultContent, ok := content.(string)
		if !ok {
			return nil, errors.New("invalid result type")
		}
		val.Content = resultContent
	}
	return &api.GetResultResponse{Item: val}, nil
}

func (k *Kobe) ListResult(ctx context.Context, req *api.ListResultRequest) (*api.ListResultResponse, error) {
	var results []*api.Result
	resultMap := k.taskCache.Items()
	for taskId := range resultMap {
		item := resultMap[taskId].Object
		val, ok := item.(*api.Result)
		if !ok {
			continue
		}
		results = append(results, val)
	}
	return &api.ListResultResponse{
		Items: results,
	}, nil
}
