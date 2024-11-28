package models

// ServerConfig contient la configuration d'un serveur
type ServerConfig struct {
	URL      string
	Username string
	Password string
}

// AgentData représente les données d'un agent.
type AgentData struct {
	LiveAgentID     string `json:"live_agent_id"`
	User            string `json:"user"`
	ServerIP        string `json:"server_ip"`
	Status          string `json:"status"`
	CampaignID      string `json:"campaign_id"`
	CallsToday      string `json:"calls_today"`
	LastCallTime    string `json:"last_call_time"`
	LastStateChange string `json:"last_state_change"`
	RowClass        string
	SourceDomain    string // Domaine source des données

}

type FetchError struct {
	Message string `json:"message"`
}

// Data contient la liste des agents.
type Data struct {
	Agents []AgentData
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
