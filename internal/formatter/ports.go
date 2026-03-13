package formatter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/docker/docker/api/types"
)

func FormatPorts(ports []types.Port) string {
	if len(ports) == 0 {
		return ""
	}

	byProto := map[string][]string{}
	for _, p := range ports {
		proto := p.Type
		if proto == "" {
			proto = "tcp"
		}

		if p.PublicPort > 0 {
			byProto[proto] = append(byProto[proto], fmt.Sprintf("%d->%d", p.PublicPort, p.PrivatePort))
		} else {
			byProto[proto] = append(byProto[proto], fmt.Sprintf("%d", p.PrivatePort))
		}
	}

	protos := make([]string, 0, len(byProto))
	for proto := range byProto {
		protos = append(protos, proto)
	}
	sort.Strings(protos)

	parts := make([]string, 0, len(protos))
	for _, proto := range protos {
		parts = append(parts, fmt.Sprintf("%s: %s", proto, strings.Join(byProto[proto], ", ")))
	}

	return strings.Join(parts, "  ")
}
