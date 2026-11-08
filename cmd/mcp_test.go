package cmd

import (
	"testing"
)

func TestHandleMCPRequestInitialize(t *testing.T) {
	req := MCPRequest{
		JsonRPC: "2.0",
		Method:  "initialize",
		ID:      1,
	}
	resp := handleMCPRequest(req, nil)
	if resp.Error != nil {
		t.Fatalf("unexpected error on initialize: %v", resp.Error)
	}
	resMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", resp.Result)
	}
	serverInfo, ok := resMap["serverInfo"].(map[string]interface{})
	if !ok || serverInfo["name"] != "auto-pr-mcp" {
		t.Fatalf("unexpected serverInfo: %v", resMap["serverInfo"])
	}
}

func TestHandleMCPRequestNotificationsInitialized(t *testing.T) {
	req := MCPRequest{
		JsonRPC: "2.0",
		Method:  "notifications/initialized",
	}
	resp := handleMCPRequest(req, nil)
	if resp.Error != nil {
		t.Fatalf("unexpected error on notifications/initialized: %v", resp.Error)
	}
}
