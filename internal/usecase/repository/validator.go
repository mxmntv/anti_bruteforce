package repository

import (
	"fmt"
	"net"

	internalError "github.com/mxmntv/anti_bruteforce/internal/errors"
)

func CheckIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("validation error: %w", internalError.ErrorInvalidIP)
	}
	return nil
}

func CheckIPNet(ip string) error {
	if _, _, err := net.ParseCIDR(ip); err != nil {
		return fmt.Errorf("validation error: %w", internalError.ErrorInvalidIPNet)
	}
	return nil
}

func CheckListName(list string) error {
	if list != "blacklist" && list != "whitelist" {
		return fmt.Errorf("validation error: %w", internalError.ErrorInvalidListName)
	}
	return nil
}
