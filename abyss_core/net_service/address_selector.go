package net_service

import (
	"errors"
	"net"
	"sync"
)

type BetaAddressSelector struct {
	localPrivateAddr net.IP
	localPublicAddr  net.IP //can be added later

	mtx *sync.Mutex
}

func NewBetaAddressSelector() (*BetaAddressSelector, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			var ip net.IP

			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip loopback and non-IPv4 addresses
			if ip == nil || ip.IsLoopback() || ip.To4() == nil || ip.IsLinkLocalUnicast() {
				continue
			}

			//fmt.Println("ffff: " + ip.String())

			return &BetaAddressSelector{
				ip,
				net.IPv4zero,
				new(sync.Mutex),
			}, nil
		}
	}

	return nil, errors.New("no network interface available")
}

func (s *BetaAddressSelector) SetPublicIP(ip net.IP) {
	s.mtx.Lock()
	s.localPublicAddr = ip
	s.mtx.Unlock()
}

func (s *BetaAddressSelector) LocalPrivateIPAddr() net.IP {
	return s.localPrivateAddr
}

func (s *BetaAddressSelector) FilterAddressCandidates(addresses []*net.UDPAddr) []*net.UDPAddr {
	public_addresses := make([]*net.UDPAddr, 0)

	var loopbackaddr *net.UDPAddr
	var privateaddr *net.UDPAddr

	for _, address := range addresses {
		if address.IP.Equal(net.IPv4zero) || address.IP.Equal(net.IPv4bcast) {
			continue
		}

		if address.IP.Equal([]byte{127, 0, 0, 1}) {
			loopbackaddr = address
			continue
		}

		if address.IP[0] == 192 && address.IP[1] == 168 {
			privateaddr = address
			continue
		}

		s.mtx.Lock()
		is_pub_eq := address.IP.Equal(s.localPublicAddr)
		s.mtx.Unlock()
		if is_pub_eq {
			continue //ignore same public address
		}

		public_addresses = append(public_addresses, address)
	}

	if len(public_addresses) == 0 { //no public address found
		if privateaddr != nil && !privateaddr.IP.Equal(s.localPrivateAddr) {
			return []*net.UDPAddr{privateaddr}
		}

		if loopbackaddr != nil {
			return []*net.UDPAddr{loopbackaddr}
		}
		return []*net.UDPAddr{}
	}
	return public_addresses
}
