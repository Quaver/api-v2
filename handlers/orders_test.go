package handlers

import "testing"

func TestGetOrderIpIpv4(t *testing.T) {
	ip := getOrderIp("192.168.1.1")

	if ip != "192.168.1.1" {
		t.Fatalf("incorrect ip, got: %v", ip)
	}
}

func TestGetOrderIpIpv6(t *testing.T) {
	ip := getOrderIp("2001:db8::1")

	if ip != "1.1.1.1" {
		t.Fatalf("incorrect ip, got: %v", ip)
	}
}

func TestGetOrderIpIpv62(t *testing.T) {
	ip := getOrderIp("2001:0db8:85a3:0000:0000:8a2e:0370:7334")

	if ip != "1.1.1.1" {
		t.Fatalf("incorrect ip, got: %v", ip)
	}
}

func TestGetOrderIpInvalid(t *testing.T) {
	ip := getOrderIp("TEST")

	if ip != "1.1.1.1" {
		t.Fatalf("incorrect ip, got: %v", ip)
	}
}
