#!/bin/bash

# Test script for MCP functionality
echo "Testing Auto PR MCP Server..."

# Test 1: Get repository status
echo -e '\nTest 1: Repository Status'
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"repo_status","arguments":{}}}' | ./auto-pr mcp 2>/dev/null | grep -A20 '"result"'

# Test 2: Analyze changes
echo -e '\nTest 2: Analyze Changes'
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"analyze_changes","arguments":{"base":"main"}}}' | ./auto-pr mcp 2>/dev/null | grep -A20 '"result"'

# Test 3: Create PR (dry run)
echo -e '\nTest 3: Create PR (dry run)'
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"create_pr","arguments":{"title":"Test PR","body":"This is a test PR","dry_run":true}}}' | ./auto-pr mcp 2>/dev/null | grep -A20 '"result"'

echo -e '\nMCP tests completed!'