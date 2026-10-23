package rediskey

import "github.com/google/uuid"

func GatewayUsers(gatewayID uuid.UUID) string {
	return "gateway_users:" + gatewayID.String()
}

func GatewayTime(gatewayID uuid.UUID) string {
	return "gateway_time:" + gatewayID.String()
}

func GatewayMaintainer() string {
	return "gateway_maintainer"
}
