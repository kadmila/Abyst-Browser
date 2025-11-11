package aurl

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

type AURL struct {
	Scheme    string
	Hash      string
	Addresses []*net.UDPAddr
	Path      string
}

func (a *AURL) ToString() string {
	if len(a.Addresses) == 0 {
		return a.Scheme + ":" + a.Hash + "/" + a.Path
	}
	candidates_string := make([]string, len(a.Addresses))
	for i, c := range a.Addresses {
		candidates_string[i] = c.String()
	}
	return a.Scheme + ":" + a.Hash + ":" + strings.Join(candidates_string, "|") + "/" + a.Path
}

const base58Chars = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func IsValidPeerID(input string) bool {
	if len(input) < 32 || input[0] < 'A' || input[0] > 'Z' {
		return false
	}
	for _, c := range input[1:] {
		if !strings.ContainsRune(base58Chars, c) {
			return false
		}
	}
	return true
}

func TryParse(input string) (*AURL, error) {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "abyss:") {
		return tryParseAbyss(input)
	} else if strings.HasPrefix(input, "abyst:") {
		return tryParseAbyst(input)
	} else {
		return &AURL{}, errors.New("invalid scheme")
	}
}

func tryParseAbyss(input string) (*AURL, error) {
	body := input[len("abyss:"):]
	if body == "" {
		return &AURL{}, errors.New("empty address")
	}

	addrStart := strings.Index(body, ":")
	pathStart := strings.Index(body, "/")

	result := &AURL{
		Scheme: "abyss",
	}

	if addrStart == -1 {
		if pathStart == -1 {
			if !IsValidPeerID(body) {
				return &AURL{}, errors.New("invalid peer hash")
			}
			result.Hash = body
			return result, nil
		}
		peerId := body[:pathStart]
		if !IsValidPeerID(peerId) {
			return &AURL{}, errors.New("invalid peer hash")
		}
		result.Hash = peerId
		result.Path = body[pathStart+1:]
		return result, nil
	}

	peerId := body[:addrStart]
	if !IsValidPeerID(peerId) {
		return &AURL{}, errors.New("invalid peer hash")
	}
	result.Hash = peerId

	var addrPart string
	if pathStart != -1 {
		addrPart = body[addrStart+1 : pathStart]
		result.Path = body[pathStart+1:]
	} else {
		addrPart = body[addrStart+1:]
	}

	result.Addresses = make([]*net.UDPAddr, 0)
	for _, ep := range strings.Split(addrPart, "|") {
		if ep == "" {
			continue
		}
		var ipPart, portPart string
		if strings.HasPrefix(ep, "[") {
			closeIdx := strings.Index(ep, "]")
			if closeIdx <= 0 {
				continue
			}
			ipPart = ep[1:closeIdx]
			if closeIdx+1 < len(ep) && ep[closeIdx+1] == ':' {
				portPart = ep[closeIdx+2:]
			} else {
				continue
			}
		} else {
			parts := strings.Split(ep, ":")
			if len(parts) != 2 {
				continue
			}
			ipPart = parts[0]
			portPart = parts[1]
		}

		port, err := strconv.Atoi(portPart)
		if net.ParseIP(ipPart) != nil && err == nil {
			result.Addresses = append(result.Addresses, &net.UDPAddr{
				IP:   net.ParseIP(ipPart),
				Port: port,
			})
		}
	}

	return result, nil
}

func tryParseAbyst(input string) (*AURL, error) {
	body := input[len("abyst:"):]

	result := &AURL{
		Scheme: "abyst",
	}

	slashIndex := strings.Index(body, "/")
	if slashIndex == -1 {
		if !IsValidPeerID(body) {
			return &AURL{}, errors.New("invalid peer hash")
		}
		result.Hash = body
		return result, nil
	}

	result.Hash = body[:slashIndex]
	result.Path = body[slashIndex+1:]
	return result, nil
}
