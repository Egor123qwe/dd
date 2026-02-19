package network

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gitlab.roy9.ru/roy9/backend/clientside/runtimedaemonv2/internal/model/network/tailscale"
)

const (
	commandPrefix = "tailscale"
	timeout       = 15 * time.Second

	maxLengthDnsValidClientID = 22
)

type Status struct {
	BackendState   string    `json:"BackendState"`
	TailscaleIPs   []string  `json:"TailscaleIPs"`
	MagicDNSSuffix string    `json:"MagicDNSSuffix"`
	Peer           peerState `json:"Peer"`

	Self struct {
		HostName string `json:"HostName"`
		Online   bool   `json:"Online"`
	} `json:"Self"`
}

type peerState map[string]struct {
	HostName string `json:"HostName"`
}

var (
	dnsSubReg = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

func (s service) version(ctx context.Context) (string, error) {
	ctx, _ = context.WithTimeout(ctx, timeout)

	data, err := s.system.Run(ctx, []string{commandPrefix, "--version"})
	if err != nil {
		return "", err
	}

	return parseVersion(string(data)), nil
}

func (s service) status(ctx context.Context) (Status, error) {
	var result Status

	data, err := s.system.Run(ctx, []string{commandPrefix, "status", "--json"})
	if err != nil {
		return Status{}, nil
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return Status{}, err
	}

	return result, nil
}

func (s service) hostname(clientID string) string {
	dnsValidClientID := strings.ToLower(
		dnsSubReg.ReplaceAllString(clientID, "-"),
	)

	hashFromSrcClientID := fmt.Sprintf("%x", md5.Sum([]byte(clientID)))

	maxLength := maxLengthDnsValidClientID

	if len(dnsValidClientID) > maxLength {
		dnsValidClientID = dnsValidClientID[:maxLength]
	}

	return fmt.Sprintf("machine-%s-%s", dnsValidClientID, hashFromSrcClientID)
}

func (s service) ip(data []string) tailscale.IP {
	if len(data) < 2 {
		return tailscale.IP{}
	}

	result := tailscale.IP{
		IpV4: data[0], // first ip  - ipv4 | in network logs
		IpV6: data[1], // second ip - ipv6 | in network logs
	}

	return result
}

func (s service) peerHostname(data peerState) []string {
	result := make([]string, 0)

	for _, v := range data {
		result = append(result, v.HostName)
	}

	return result
}

func parseVersion(str string) string {
	lines := strings.Split(str, "\n")

	for _, line := range lines {
		if line != "" {
			return line
		}
	}

	return ""
}
