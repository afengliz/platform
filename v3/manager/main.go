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
	pid        string
	cpu        float64
	memory     float64
	cpuTime    int
	memoryTime int
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
	agents sync.Map // map[string]*Agent // key agentID

	pluginRuntimes sync.Map // map[string]*PluginPod // key instanceID + version

	pluginMetrics sync.Map //map[string]*PluginMetric // key instanceID + version
}

// push 任务
func (m *Manager) PushTask(task *Task) {
	m.taskQueue <- task
}

func (m *Manager) Report(agent Agent, items map[string]PluginReportItem) {
	// agent 更新
	if old, ok := m.agents.LoadOrStore(agent.agentID, agent); ok {
		m.agents.CompareAndSwap(agent.agentID, old, agent)
	}

	for _, item := range items { // 多的部分
		// 更新 plugin runtime
		if val, ok := m.pluginRuntimes.LoadOrStore(item.instanceID, PluginPod{
			instanceID: item.instanceID,
			version:    item.version,
			agentID:    agent.agentID,
			agentIP:    agent.agentIP,
		}); ok { // 存在
			old := val.(PluginPod)
			if old.agentID == agent.agentID { // 同一个 agent
				old.runtimeStatus = "running"
				old.lastTimeStamp = time.Now().Second()
			} else { // 不同 agent
				//  todo 不合理的数据，直接调用agent的删除pod接口，可以异步
			}
		}
		// 更新metric
		if val, ok := m.pluginMetrics.LoadOrStore(item.instanceID, PluginMetric{
			pid:     item.instanceID,
			cpu:     item.cpu,
			memory:  item.memory,
			cpuTime: time.Now().Second(),
		}); ok {
			oldO := val.(PluginMetric)
			newO := oldO
			if oldO.cpu < item.cpu {
				newO.cpuTime = item.cpuTime
				newO.cpu = item.cpu
			}
			if oldO.memory < item.memory {
				newO.memoryTime = item.memoryTime
				newO.memory = item.memory
			}
			m.pluginMetrics.CompareAndSwap(item.instanceID, val, newO)
		}
	}
	// todo 貌似不需要了
	m.pluginRuntimes.Range(func(key, value any) bool {
		aPod := value.(PluginPod)
		if aPod.agentID == agent.agentID {
			if _, ok := items[aPod.instanceID]; !ok {
				// todo 少的部分，将插件运行时状态，标记为killed
			}
		}
		return true
	})

}

func (m *Manager) UnRegisterAgent(agentID string) {
	m.agents.Delete(agentID)
}

// 调度
func (m *Manager) Schedule() {
	for {
		select {
		case task := <-m.taskQueue:
			if task.Action == "start" { // 启动
				if val, ok := m.pluginRuntimes.LoadOrStore(task.InstanceID, PluginPod{
					instanceID:    task.InstanceID,
					version:       task.Version,
					runtimeStatus: "pending",
				}); ok {
					pod := val.(PluginPod)
					if pod.runtimeStatus == "running" {
						continue
					}
					if pod.runtimeStatus == "pending" {
						continue
					}
				}
				// 计算最合适的 agent（哪个插件少，就漂哪个）

				// agent plugin 数量 + 1

				// push to agent
			} else { //停止
				m.pluginRuntimes.Delete(task.InstanceID)
				// push to agent
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
		// 获取当前业务状态为启用的插件列表，与当前运行时列表做对比
		// 少的部分push add task，多的部分push stop task
	}
}

// 监控Agent的心跳长时间无上报的，将Agent上的运行时在内存里全部删除
func (m *Manager) Monitor3() {
	for {
		m.lock.RLock()
		// 拷贝出一份agent列表
		m.lock.RUnlock()
		// 判断是否有Agent的心跳长时间无上报的，有的话再上写锁
		// 将分配在具体Agent上的运行时在内存里删除
		// 释放写锁
	}
}
