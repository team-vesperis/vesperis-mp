package database

import (
	"strings"

	"go.minekube.com/gate/pkg/util/uuid"
)

func redisPlayerKeyTranslator(playerId uuid.UUID) string {
	return "player_data:" + playerId.String()
}

func redisPartyKeyTranslator(partyId uuid.UUID) string {
	return "party_data:" + partyId.String()
}

func redisProxyKeyTranslator(proxyId uuid.UUID) string {
	return "proxy_data:" + proxyId.String()
}

func redisBackendKeyTranslator(backendId uuid.UUID) string {
	return "backend_data:" + backendId.String()
}

func safeJsonPathForPostgres(field string) string {
	parts := strings.Split(field, ".")
	for i, part := range parts {
		if strings.ContainsAny(part, ` ."`) {
			parts[i] = `"` + strings.ReplaceAll(part, `"`, `\"`) + `"`
		}
	}
	return "{" + strings.Join(parts, ",") + "}"
}
