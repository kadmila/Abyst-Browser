package aurl

// import (
// 	"net"
// 	"net/netip"
// 	"strconv"
// 	"strings"
// )

// type AURL struct {
// 	Scheme    string //abyss(ahmp) or abyst(http3)
// 	Hash      string
// 	Addresses []*net.UDPAddr
// 	Path      string
// }

// // abyss:abc:9.8.7.6:1605/somepath
// // abyss:abc:[2001:db8:85a3:8d3:1319:8a2e:370:7348]:443|9.8.7.6:1605/somepath
// // abyss:abc/somepath
// // abyss:abc:9.8.7.6:1605
// // abyss:abc
// func (a *AURL) ToString() string {
// 	if len(a.Addresses) == 0 {
// 		return "abyss:" + a.Hash + a.Path
// 	}
// 	candidates_string := make([]string, len(a.Addresses))
// 	for i, c := range a.Addresses {
// 		candidates_string[i] = c.String()
// 	}
// 	return "abyss:" + a.Hash + ":" + strings.Join(candidates_string, "|") + a.Path
// }

// type AURLParseError struct {
// 	Code int
// }

// func (u *AURLParseError) Error() string {
// 	var msg string
// 	switch u.Code {
// 	case 100:
// 		msg = "unsupported protocol"
// 	case 101:
// 		msg = "invalid format"
// 	case 102:
// 		msg = "hash too short"
// 	case 103:
// 		msg = "address candidate parse fail"
// 	default:
// 		msg = "unknown error (" + strconv.Itoa(u.Code) + ")"
// 	}
// 	return "failed to parse abyssURL: " + msg
// }

// func ParseAURL(raw string) (*AURL, error) {
// 	var scheme string
// 	var body string
// 	var ok bool
// 	if body, ok = strings.CutPrefix(raw, "abyss:"); ok {
// 		scheme = "abyss"
// 	} else {
// 		if body, ok = strings.CutPrefix(raw, "abyst:"); ok {
// 			scheme = "abyst"
// 		} else {
// 			return nil, &AURLParseError{Code: 100}
// 		}
// 	}

// 	body = strings.TrimPrefix(body, "//") //for old people

// 	hash_endpos := strings.IndexAny(body, ":/")
// 	if hash_endpos == -1 {
// 		//no candidates, no path
// 		return &AURL{
// 			Scheme:    scheme,
// 			Hash:      body,
// 			Addresses: []*net.UDPAddr{},
// 			Path:      "/",
// 		}, nil
// 	}
// 	hash := body[:hash_endpos]
// 	if len(hash) < 1 {
// 		return nil, &AURLParseError{Code: 102}
// 	}

// 	if body[hash_endpos] == ':' {
// 		cand_path := body[hash_endpos+1:]

// 		pathpos := strings.Index(cand_path, "/")
// 		var candidates_str string
// 		var path string
// 		if pathpos == -1 {
// 			//no path
// 			candidates_str = cand_path
// 			path = "/"
// 		} else {
// 			candidates_str = cand_path[:pathpos]
// 			path = cand_path[pathpos:]
// 		}

// 		c_split := strings.Split(candidates_str, "|")
// 		candidates := make([]*net.UDPAddr, len(c_split))
// 		for i, candidate := range c_split {
// 			addrport, err := netip.ParseAddrPort(candidate)
// 			if err != nil {
// 				return nil, err
// 			}
// 			cand := net.UDPAddrFromAddrPort(addrport)
// 			if cand.Port == 0 {
// 				return nil, &AURLParseError{Code: 103}
// 			}
// 			candidates[i] = cand
// 		}

// 		return &AURL{
// 			Scheme:    scheme,
// 			Hash:      hash,
// 			Addresses: candidates,
// 			Path:      path,
// 		}, nil
// 	} else if body[hash_endpos] == '/' {
// 		//only path
// 		return &AURL{
// 			Scheme:    scheme,
// 			Hash:      hash,
// 			Addresses: []*net.UDPAddr{},
// 			Path:      body[hash_endpos:],
// 		}, nil
// 	}
// 	panic("ParseAbyssURL: implementation error")
// }
