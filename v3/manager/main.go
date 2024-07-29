package manager

import (
	"sync"
)

type PluginRuntime struct {
	status     string
	instanceID string
	version    string
	agentID    string
	agentIP    string
}

type PluginRuntimeMetric struct {
	instanceID string
	pid        string
	cpu        float64
	memory     float64
	cpuTime    int
	memoryTime int
}

type Agent struct {
	agentID string
	agentIP string
}

type AgentMetric struct {
	cpu           float64
	memory        float64
	usageCpu      float64
	usageMemory   float64
	lastTimestamp int
}

type Task struct {
	InstanceID string
	Version    string
	Action     string
	Retry      int
}

type Manager struct {
	// 调度任务队列
	taskQueue chan *Task
	// agent队列锁
	agentQueueRW sync.RWMutex
	// agent队列
	agentQueue map[string]chan *Task
	// 通过注册/注销 + 心跳保活（超过驱逐）
	agents sync.Map // map[string]*Agent // key agentID
	// agent metric
	agentMetrics sync.Map // map[string]*AgentMetric // key agentID
	// agent 预调度的Plugin数量
	agentPreSchedulePluginsCount sync.Map // map[string]int // key agentID,val plugin count
	// 插件运行时
	pluginRuntimes sync.Map // map[string]*PluginRuntime // key instanceID + version
	// 插件运行时metric
	pluginRuntimeMetric sync.Map //map[string]*PluginRuntimeMetric // key instanceID + version
}

// push 任务
func (m *Manager) PushTask(task *Task) {
	m.taskQueue <- task
}

func (m *Manager) ReportRunningPlugins(agentID string, runningPlugins map[string]struct{}) {
	if _, ok := m.agents.Load(agentID); !ok { // 不存在，直接返回
		return
	}
	// runtime status 更新

	// 这部分，放agent自己处理可能更好些
	// 比runtime多的，push stop task
	// 比runtime少的，，push start task
}
func (m *Manager) ReportMetric(agentID string, agentMetric AgentMetric, runningPluginMetrics map[string]PluginRuntimeMetric) {
	if _, ok := m.agents.Load(agentID); !ok { // 不存在，直接返回
		return
	}
	// agent metric 更新

	// runtime metric 更新
}

// 注册agent
func (m *Manager) RegisterAgent(agentID string, agentIP string) {
	m.agents.Store(agentID, Agent{
		agentID: agentID,
		agentIP: agentIP,
	})
}

// 注销agent
func (m *Manager) UnRegisterAgent(agentID string) {
	m.agents.Delete(agentID)
}

// 调度
func (m *Manager) Schedule() {
	for {
		select {
		case task := <-m.taskQueue:
			if task.Action == "start" { // 启动
				if val, ok := m.pluginRuntimes.LoadOrStore(task.InstanceID, PluginRuntime{
					instanceID: task.InstanceID,
					version:    task.Version,
					status:     "pending",
				}); ok {
					pod := val.(PluginRuntime)
					if pod.status == "running" {
						continue
					}
					if pod.status == "pending" {
						continue
					}
					if pod.status == "pushed" {
						continue
					}
				}
				// 计算最合适的 agent（哪个插件少，就漂哪个）

				// push to agent
				// pod.agentID = agentID
				// pod.agentIP = agentIP
				// pod.status = "pushed"
				// m.pluginRuntimes.CompareAndSwap(task.InstanceID, val, pod)
				// 失败推送至队列尾部
			} else { //停止
				m.pluginRuntimes.Delete(task.InstanceID)

				// push to agent
			}
		}
	}
}

// 更新运行时状态
func (m *Manager) UpdateRuntimeStatus(instanceID string, status string) {
	if val, ok := m.pluginRuntimes.Load(instanceID); ok {
		pod := val.(PluginRuntime)
		pod.status = status
		m.pluginRuntimes.CompareAndSwap(instanceID, val, pod)
	}
}

// 监控runtime metric 的插件运行时的lasttimestamp，超过一定时间没有心跳，就认为插件挂了，直接删除内存runtime
func (m *Manager) Monitor1() {
	for {
		// 直接删除内存runtime
	}
}

// 监控当前运行时与业务状态的一致性
func (m *Manager) Monitor2() {
	for {
		// 获取当前业务状态为启用的插件列表，与当前运行时列表做对比
		// 少的部分push add task，多的部分push stop task
	}
}
