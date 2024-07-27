package manager

import (
	"context"
	"encoding/json"
	"time"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
)

type PluginPod struct {
	resourceVersion int
	runtimeStatus   string
	instanceID      string
	version         string
	agentID         string
	agentIP         string
	lastTimeStamp   int
}

type PluginMetric struct {
	pid    string
	cpu    float64
	memory float64
	time   int
}
type PluginReportItem struct {
	instanceID      string
	version         string
	resourceVersion int
	PluginMetric
}

type Agent struct {
	agentID        string
	agentIP        string
	cpu            float64
	memory         float64
	usageCpu       float64
	usageMemory    float64
	runningPlugins []string
	lastTimestamp  int
}
type Task struct {
	InstanceID string
	Version    string
	Action     string
}

type Manager struct {
	taskQueue  chan *Task
	etcdClient *clientv3.Client
	ctx        context.Context
}

func NewManager(etcdEndpoints []string) (*Manager, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &Manager{
		taskQueue:  make(chan *Task, 100),
		etcdClient: cli,
		ctx:        context.Background(),
	}, nil
}

func (m *Manager) PushTask(task *Task) {
	m.taskQueue <- task
}

func (m *Manager) Report(agent *Agent, items map[string]PluginReportItem) error {
	s, err := concurrency.NewSession(m.etcdClient)
	if err != nil {
		return err
	}
	defer s.Close()
	mu := concurrency.NewMutex(s, "/locks/report")

	if err := mu.Lock(m.ctx); err != nil {
		return err
	}
	defer mu.Unlock(m.ctx)

	// 更新 agent 状态
	agentKey := "/agents/" + agent.agentID
	agentValue, _ := json.Marshal(agent)
	_, err = m.etcdClient.Put(m.ctx, agentKey, string(agentValue))
	if err != nil {
		return err
	}

	// 更新 plugin runtimes 和 metrics
	for _, item := range items {
		runtimeKey := "/pluginRuntimes/" + item.instanceID
		runtimeValue, _ := json.Marshal(&PluginPod{
			instanceID:    item.instanceID,
			version:       item.version,
			agentID:       agent.agentID,
			agentIP:       agent.agentIP,
			lastTimeStamp: time.Now().Second(),
			runtimeStatus: "running",
		})
		_, err = m.etcdClient.Put(m.ctx, runtimeKey, string(runtimeValue))
		if err != nil {
			return err
		}

		metricKey := "/pluginMetrics/" + item.instanceID
		metricValue, _ := json.Marshal(&PluginMetric{
			pid:    item.instanceID,
			cpu:    item.cpu,
			memory: item.memory,
			time:   time.Now().Second(),
		})
		_, err = m.etcdClient.Put(m.ctx, metricKey, string(metricValue))
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) UnRegisterAgent(agentID string) error {
	s, err := concurrency.NewSession(m.etcdClient)
	if err != nil {
		return err
	}
	defer s.Close()
	mu := concurrency.NewMutex(s, "/locks/unregister")

	if err := mu.Lock(m.ctx); err != nil {
		return err
	}
	defer mu.Unlock(m.ctx)

	agentKey := "/agents/" + agentID
	_, err = m.etcdClient.Delete(m.ctx, agentKey)
	return err
}

func (m *Manager) Schedule() {
	for {
		select {
		case task := <-m.taskQueue:
			runtimeKey := "/pluginRuntimes/" + task.InstanceID
			resp, err := m.etcdClient.Get(m.ctx, runtimeKey)
			if err != nil || len(resp.Kvs) == 0 {
				continue
			}

			var pluginPod PluginPod
			json.Unmarshal(resp.Kvs[0].Value, &pluginPod)

			if task.Action == "start" {
				if pluginPod.runtimeStatus == "running" || pluginPod.runtimeStatus == "pending" {
					continue
				}

				pluginPod.runtimeStatus = "pending"
				runtimeValue, _ := json.Marshal(pluginPod)
				m.etcdClient.Put(m.ctx, runtimeKey, string(runtimeValue))

				// 计算最合适的 agent，并推送到 agent
			} else {
				if pluginPod.runtimeStatus == "killing" {
					continue
				}

				pluginPod.runtimeStatus = "killing"
				runtimeValue, _ := json.Marshal(pluginPod)
				m.etcdClient.Put(m.ctx, runtimeKey, string(runtimeValue))

				// 推送到 agent
			}
		}
	}
}

func (m *Manager) Monitor1() {
	for {
		resp, err := m.etcdClient.Get(m.ctx, "/pluginRuntimes", clientv3.WithPrefix())
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		for _, kv := range resp.Kvs {
			var pluginPod PluginPod
			json.Unmarshal(kv.Value, &pluginPod)

			if time.Now().Second()-pluginPod.lastTimeStamp > 30 {
				m.PushTask(&Task{
					InstanceID: pluginPod.instanceID,
					Version:    pluginPod.version,
					Action:     "stop",
				})
			}
		}

		time.Sleep(10 * time.Second)
	}
}

func (m *Manager) Monitor2() {
	for {
		// 获取当前业务状态为启用的插件列表，与当前运行时列表做对比
		// 少的部分 push add task，多的部分 push stop task

		time.Sleep(30 * time.Second)
	}
}

func (m *Manager) Monitor3() {
	for {
		resp, err := m.etcdClient.Get(m.ctx, "/agents", clientv3.WithPrefix())
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		for _, kv := range resp.Kvs {
			var agent Agent
			json.Unmarshal(kv.Value, &agent)

			if time.Now().Second()-agent.lastTimestamp > 60 {
				s, err := concurrency.NewSession(m.etcdClient)
				if err != nil {
					continue
				}
				defer s.Close()
				mu := concurrency.NewMutex(s, "/locks/agent_cleanup")

				if err := mu.Lock(m.ctx); err != nil {
					continue
				}

				m.etcdClient.Delete(m.ctx, "/agents/"+agent.agentID)
				mu.Unlock(m.ctx)
			}
		}

		time.Sleep(10 * time.Second)
	}
}
