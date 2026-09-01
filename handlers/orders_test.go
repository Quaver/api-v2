package handlers

import (
	"encoding/json"
	"github.com/Quaver/api2/db"
	"testing"
)

func TestUserOrderResponseIncludesPaymentIdentifiers(t *testing.T) {
	order := &db.Order{
		OrderId:       123,
		TransactionId: "transaction-456",
		IPAddress:     "192.168.1.1",
	}

	data, err := json.Marshal(userOrderResponse{
		Order:         order,
		OrderId:       order.OrderId,
		TransactionId: order.TransactionId,
	})
	if err != nil {
		t.Fatal(err)
	}

	var response map[string]any
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}

	if response["order_id"] != float64(123) {
		t.Fatalf("expected order_id to be included, got %#v", response["order_id"])
	}

	if response["transaction_id"] != "transaction-456" {
		t.Fatalf("expected transaction_id to be included, got %#v", response["transaction_id"])
	}

	if _, exists := response["ip_address"]; exists {
		t.Fatal("expected ip_address to remain private")
	}
}

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
