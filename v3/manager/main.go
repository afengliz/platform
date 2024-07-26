package manager

import (
	"sync"
	"time"
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
	lock      sync.RWMutex // 读写锁,todo 会不会有死锁问题
	taskQueue chan *Task
	// 通过注册/注销 + 心跳保活（超过驱逐）
	agents map[string]*Agent // key agentID

	pluginRuntimes map[string]*PluginPod // key instanceID + version

	pluginMetrics map[string]*PluginMetric // key instanceID + version
}

// push 任务
func (m *Manager) PushTask(task *Task) {
	m.taskQueue <- task
}

func (m *Manager) Report(agent *Agent, items map[string]PluginReportItem) {
	m.agents[agent.agentID] = agent
	agent.lastTimestamp = time.Now().Second()
	// todo 指标部分，可以异步更新到其他对象，减少锁的时间
	for _, item := range items { // 多的部分
		// 更新 plugin runtime
		if m.pluginRuntimes[item.instanceID] != nil { // 存在
			if agent.agentID == m.pluginRuntimes[item.instanceID].agentID { // 同一个 agent
				m.pluginRuntimes[item.instanceID].runtimeStatus = "running"
				m.pluginRuntimes[item.instanceID].lastTimeStamp = time.Now().Second()
			} else { // 不同 agent
				// todo 挪动到最后一步 (直接调用，从原来的 agent 中删除)
			}
		} else { // 不存在
			m.pluginRuntimes[item.instanceID] = &PluginPod{
				instanceID: item.instanceID,
				version:    item.version,
				agentID:    agent.agentID,
				agentIP:    agent.agentIP,
			}
		}
		if m.pluginMetrics[item.instanceID] != nil {
			// 只保留最大值
		} else {
			m.pluginMetrics[item.instanceID] = &PluginMetric{
				pid:    item.instanceID,
				cpu:    item.cpu,
				memory: item.memory,
				time:   time.Now().Second(),
			}
		}
	}
	// todo 少的部分，将插件运行时状态，标记为killed

}
func (m *Manager) UnRegisterAgent(agentID string) {
	delete(m.agents, agentID)
}

// 调度
func (m *Manager) Schedule() {
	for {
		select {
		case task := <-m.taskQueue:
			if task.Action == "start" { // 启动
				if m.pluginRuntimes[task.InstanceID] != nil {
					if m.pluginRuntimes[task.InstanceID].runtimeStatus == "running" {
						continue
					}
					if m.pluginRuntimes[task.InstanceID].runtimeStatus == "pending" {
						continue
					}
				} else {
					m.pluginRuntimes[task.InstanceID] = &PluginPod{
						instanceID: task.InstanceID,
						version:    task.Version,
					}
				}
				m.pluginRuntimes[task.InstanceID].runtimeStatus = "pending"
				// 计算最合适的 agent（哪个插件少，就漂哪个）

				// agent plugin 数量 + 1

				// push to agent
			} else { //停止
				if _, ok := m.pluginRuntimes[task.InstanceID+task.Version]; ok {
					m.pluginRuntimes[task.InstanceID+task.Version].runtimeStatus = "killing"

					// push to agent
				}
			}
		}
	}
}

// 监控running状态的插件运行时的lasttimestamp，超过一定时间没有心跳，就认为插件挂了，直接给push 一个停止任务
func (m *Manager) Monitor1() {
	for {
		m.lock.RLock()
		// 拷贝出一份 超过一定时间没有心跳的runtime
		m.lock.RUnlock()
		// stop‘s task push to  taskQueue
	}
}

// 监控当前运行时与业务状态的一致性
func (m *Manager) Monitor2() {
	for {
		m.lock.RLock()
		// 拷贝出一份 超过一定时间没有心跳的runtime
		m.lock.RUnlock()
		// 少的部分push add task，多的部分push stop task
	}
}

// 标记插件运行时为running

// 标记插件运行时为killed
