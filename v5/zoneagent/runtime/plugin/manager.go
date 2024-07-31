package plugin

import "code/platform/v5/lock"

type Plugin struct {
	instanceID string
	version    string
	appID      string
	appName    string
	agentID    string
	agentIP    string
	status     string
	lastUpdate int64
}

type Manager struct {
	lock   lock.Table
	Agents map[string]Plugin
}
