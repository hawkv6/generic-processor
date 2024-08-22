package message

type TelemetryMessage struct {
	BaseResultMessage
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
	IPv6         string `mapstructure:"state/ip" validate:"required"`
	Origin       string `mapstructure:"state/origin" validate:"required"`
	PrefixLength int    `mapstructure:"state/prefix_length" validate:"required"`
	Status       string `mapstructure:"state/status" validate:"required"`
}

type InterfaceStatusMessage struct {
	TelemetryMessage
	Fields InterfaceStatusFields
}

type InterfaceStatusFields struct {
	AdminStatus             string `mapstructure:"admin_status" validate:"required"`
	CarrierTransitionsCount int    `mapstructure:"counters/carrier_transitions" validate:"min=0"`
	CPU                     string `mapstructure:"cpu" validate:"required"`
	Delete                  bool   `mapstructure:"delete"`
	Description             string `mapstructure:"description" validate:"required"`
	Enabled                 string `mapstructure:"enabled" validate:"required"`
	InterfaceIndex          int    `mapstructure:"ifindex" validate:"min=0"`
	LastChange              int64  `mapstructure:"last_change" validate:"required"`
	Logical                 string `mapstructure:"logical" validate:"required"`
	LoopbackMode            string `mapstructure:"loopback_mode" validate:"required"`
	Management              string `mapstructure:"management" validate:"required"`
	MTU                     int    `mapstructure:"mtu" validate:"required"`
	Name                    string `mapstructure:"name" validate:"required"`
	OperStatus              string `mapstructure:"oper_status" validate:"required"`
	Type                    string `mapstructure:"type" validate:"required"`
}
