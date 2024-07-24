package message

import "github.com/jalapeno-api-gateway/jagw/pkg/arango"

type UpdateLinkMessage struct {
	arango.LSLink
	NormalizedUnidirLinkDelay      float64 `json:"normalized_unidir_link_delay,omitempty"`
	NormalizedUnidirDelayVariation float64 `json:"normalized_unidir_delay_variation,omitempty"`
	NormalizedUnidirPacketLoss     float64 `json:"normalized_unidir_packet_loss,omitempty"`
	UnidirPacketLossPercentage     float64 `json:"undir_packet_loss_percentage,omitempty"`
}

func (updateMessage *UpdateLinkMessage) GetFields() map[string]interface{} {
	return map[string]interface{}{
		"unidir_link_delay":                 &updateMessage.UnidirLinkDelay,
		"unidir_link_delay_min_max[0]":      &updateMessage.UnidirLinkDelayMinMax[0],
		"unidir_link_delay_min_max[1]":      &updateMessage.UnidirLinkDelayMinMax[1],
		"unidir_delay_variation":            &updateMessage.UnidirDelayVariation,
		"unidir_packet_loss_percentage":     &updateMessage.UnidirPacketLossPercentage,
		"max_link_bw_kbps":                  &updateMessage.MaxLinkBWKbps,
		"unidir_available_bw":               &updateMessage.UnidirAvailableBW,
		"unidir_bw_utilization":             &updateMessage.UnidirBWUtilization,
		"normalized_unidir_link_delay":      &updateMessage.NormalizedUnidirLinkDelay,
		"normalized_unidir_delay_variation": &updateMessage.NormalizedUnidirDelayVariation,
		"normalized_unidir_packet_loss":     &updateMessage.NormalizedUnidirPacketLoss,
	}
}
