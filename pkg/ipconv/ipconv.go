package ipconv

import (
	"encoding/binary"
	"errors"
	"math"
	"net"
)

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func IPv4ToInt(IPv4String string) (uint32, error) {
	IPv4Addr := net.ParseIP(IPv4String)
	if IPv4Addr == nil {
		return 0, errors.New("invalid IPv4 address")
	}
	IPv4Addr = IPv4Addr.To4()
	return binary.BigEndian.Uint32(IPv4Addr), nil
}

func IntToIPv4(IPv4Int uint32) string {
	ipBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ipBytes, IPv4Int)
	ip := net.IP(ipBytes)
	return ip.String()
}

func RoundDecIP(IPv4Int uint32, roundTo uint32) uint32 {
	roundedValue := math.Floor(float64(IPv4Int) / float64(roundTo))
	return uint32(roundedValue)
}

func RoundIP(IPv4String string, roundTo uint32) (uint32, error) {
	IPv4Int, err := IPv4ToInt(IPv4String)
	if err != nil {
		return 0, err
	}

	return RoundDecIP(IPv4Int, roundTo), nil
}

func CIDRMinMaxInt(cidr string) (uint32, uint32, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, 0, err
	}

	ips := []net.IP{ip}
	for currentIP := ip.Mask(ipnet.Mask); ipnet.Contains(currentIP); incIP(currentIP) {
		ips = append(ips, currentIP)
	}

	min, err := IPv4ToInt(ips[0].String())
	if err != nil {
		return 0, 0, err
	}

	max, err := IPv4ToInt(ips[len(ips)-1].String())
	if err != nil {
		return 0, 0, err
	}

	return min, max, nil
}
