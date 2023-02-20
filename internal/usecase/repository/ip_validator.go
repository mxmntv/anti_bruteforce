package repository

import (
	"fmt"
	"net"
)

type BadIPError struct {
	IP string
}

func (e *BadIPError) Error() string {
	return fmt.Sprintf("invalid ip address: %s", e.IP)
}

type BadIPNetError struct {
	IP string
}

func (e *BadIPNetError) Error() string {
	return fmt.Sprintf("invalid ip/mask address: %s", e.IP)
}

func CheckIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("validate ip: %w", &BadIPError{ip})
	}
	return nil
}

func CheckIPNet(ip string) error {
	if _, _, err := net.ParseCIDR(ip); err != nil {
		return fmt.Errorf("validate ip/mask: %w", &BadIPNetError{ip})
	}
	return nil
}
