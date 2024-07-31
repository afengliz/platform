package agent

import (
	"code/platform/v5/lock"
)

type Agent struct {
	agentID      string
	agentIP      string
	agentPlugins map[string]struct{}
}

type Manager struct {
	lock   lock.Table
	Agents map[string]Agent
}

func (*Manager) Register(agent Agent) {}

func (*Manager) UnRegister(agentID string) {}

func (*Manager) AddPlugin(agentID string, plugin string) error {
	panic("implement me")
}
func (*Manager) RemovePlugin(agentID string, plugin string) error {
	panic("implement me")
}
