package message

type TelemetryMessage struct {
	BaseResultmessage
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Tags      MessageTags            `json:"tags,omitempty"`
	Timestamp int64                  `json:"timestamp,omitempty"`
}

type MessageTags struct {
	Host          string `json:"host,omitempty"`
	Index         string `json:"index"`
	IP            string `json:"ip"`
	InterfaceName string `json:"name,omitempty"`
	Node          string `json:"node"`
	Path          string `json:"path,omitempty"`
	Source        string `json:"source,omitempty"`
	Subscription  string `json:"subscription,omitempty"`
}

type IPv6Message struct {
	TelemetryMessage
	Fields IPv6Fields
}

type IPv6Fields struct {
	Delete       bool   `mapstructure:"delete"`
	IPv6         string `mapstructure:"state/ip"`
	Origin       string `mapstructure:"state/origin"`
	PrefixLength int    `mapstructure:"state/prefix_length"`
	Status       string `mapstructure:"state/status"`
}

type InterfaceStatusMessage struct {
	TelemetryMessage
	Fields InterfaceStatusFields
}

type InterfaceStatusFields struct {
	AdminStatus            string `mapstructure:"admin_status"`
	CarrierTranstionsCount int    `mapstructure:"counters/carrier_transitions"`
	CPU                    string `mapstructure:"cpu"`
	Delete                 bool   `mapstructure:"delete"`
	Description            string `mapstructure:"description"`
	Enabled                string `mapstructure:"enabled"`
	InterfaceIndex         int    `mapstructure:"ifindex"`
	LastChange             int64  `mapstructure:"last_change"`
	Logical                string `mapstructure:"logical"`
	LoopbackMode           string `mapstructure:"loopback_mode"`
	Management             string `mapstructure:"management"`
	MTU                    int    `mapstructure:"mtu"`
	Name                   string `mapstructure:"name"`
	OperStatus             string `mapstructure:"oper_status"`
	Type                   string `mapstructure:"type"`
}
