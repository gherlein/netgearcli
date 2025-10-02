#!/bin/bash
# Integration test for POE management
# Tests status, enable/disable commands, and state restoration
#
# Usage: ./test-poe-toggle.sh <switch-hostname>
#
# Requirements:
#   - NETGEAR_SWITCHES environment variable must be set
#   - poe-management binary must be built in bin/
#
# This test will:
#   1. Get current POE status for all ports
#   2. Toggle each port (on→off, off→on)
#   3. Verify the changes took effect
#   4. Restore original state

set -e  # Exit on any error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SWITCH="${1:-}"
BIN="./bin/poe-management"
TMPDIR="${TMPDIR:-/tmp}"
STATE_FILE="${TMPDIR}/poe-test-state-$$.json"
NEW_STATE_FILE="${TMPDIR}/poe-test-newstate-$$.json"

# Cleanup function
cleanup() {
    rm -f "$STATE_FILE" "$NEW_STATE_FILE"
}
trap cleanup EXIT

# Usage check
if [ -z "$SWITCH" ]; then
    echo "Usage: $0 <switch-hostname>"
    echo ""
    echo "Example:"
    echo "  export NETGEAR_SWITCHES=\"tswitch16:password\""
    echo "  $0 tswitch16"
    exit 1
fi

# Check if binary exists
if [ ! -x "$BIN" ]; then
    echo -e "${RED}Error: $BIN not found or not executable${NC}"
    echo "Run 'make build' first"
    exit 1
fi

# Check if environment variable is set
if [ -z "$NETGEAR_SWITCHES" ]; then
    echo -e "${RED}Error: NETGEAR_SWITCHES environment variable not set${NC}"
    echo "Example: export NETGEAR_SWITCHES=\"tswitch16:password\""
    exit 1
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}POE Management Integration Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo "Switch: $SWITCH"
echo ""

# Step 1: Get initial state
echo -e "${YELLOW}[1/5] Getting initial POE status...${NC}"
$BIN "$SWITCH" status > "$STATE_FILE"

if [ ! -s "$STATE_FILE" ]; then
    echo -e "${RED}Failed to get POE status${NC}"
    exit 1
fi

# Parse the JSON to extract port states
# Format: {"poe_status": [{"Port ID":"1","Status":"Delivering Power",...}, ...]}
# Using grep/sed instead of jq for better portability

# Extract all port entries and parse them
PORTS_ON=""
PORTS_OFF=""

# Read each port block
grep -o '"Port ID":[^,}]*' "$STATE_FILE" | while IFS=':' read key value; do
    PORT=$(echo "$value" | tr -d ' "')

    # Find the status for this port by looking at the next few lines after Port ID
    # Extract the object containing this port
    PORT_BLOCK=$(grep -A 10 "\"Port ID\".*\"$PORT\"" "$STATE_FILE" | head -12)

    # Check if Status is "Delivering Power"
    if echo "$PORT_BLOCK" | grep -q '"Status".*"Delivering Power"'; then
        echo "$PORT" >> /tmp/ports_on_$$
    else
        echo "$PORT" >> /tmp/ports_off_$$
    fi
done

if [ -f /tmp/ports_on_$$ ]; then
    PORTS_ON=$(cat /tmp/ports_on_$$ | sort -n)
    rm -f /tmp/ports_on_$$
fi

if [ -f /tmp/ports_off_$$ ]; then
    PORTS_OFF=$(cat /tmp/ports_off_$$ | sort -n)
    rm -f /tmp/ports_off_$$
fi

PORTS_ON_ARRAY=($PORTS_ON)
PORTS_OFF_ARRAY=($PORTS_OFF)

echo -e "${GREEN}✓ Initial state captured${NC}"
echo "  Ports ON:  ${PORTS_ON_ARRAY[@]:-none}"
echo "  Ports OFF: ${PORTS_OFF_ARRAY[@]:-none}"
echo ""

# Step 2: Toggle all ports
echo -e "${YELLOW}[2/5] Toggling all ports...${NC}"

# Disable ports that are currently ON
if [ ${#PORTS_ON_ARRAY[@]} -gt 0 ]; then
    echo "  Disabling ports: ${PORTS_ON_ARRAY[@]}"
    $BIN "$SWITCH" disable ${PORTS_ON_ARRAY[@]} > /dev/null
fi

# Enable ports that are currently OFF
if [ ${#PORTS_OFF_ARRAY[@]} -gt 0 ]; then
    echo "  Enabling ports: ${PORTS_OFF_ARRAY[@]}"
    $BIN "$SWITCH" enable ${PORTS_OFF_ARRAY[@]} > /dev/null
fi

echo -e "${GREEN}✓ Ports toggled${NC}"
echo ""

# Give the switch a moment to process the changes
echo -e "${YELLOW}[3/5] Waiting 3 seconds for switch to process changes...${NC}"
sleep 3
echo ""

# Step 3: Verify the changes
echo -e "${YELLOW}[4/5] Verifying state changes...${NC}"
$BIN "$SWITCH" settings > "$NEW_STATE_FILE"

# Check that previously ON ports are now disabled
VERIFY_FAILED=0
for port in ${PORTS_ON_ARRAY[@]:-}; do
    PORT_BLOCK=$(grep -A 10 "\"Port ID\".*\"$port\"" "$NEW_STATE_FILE" | head -12)
    if echo "$PORT_BLOCK" | grep -q '"Port Power".*"disabled"'; then
        echo -e "${GREEN}✓ Port $port is disabled (was enabled)${NC}"
    else
        echo -e "${RED}✗ Port $port should be disabled but is still enabled${NC}"
        VERIFY_FAILED=1
    fi
done

# Check that previously OFF ports are now enabled
for port in ${PORTS_OFF_ARRAY[@]:-}; do
    PORT_BLOCK=$(grep -A 10 "\"Port ID\".*\"$port\"" "$NEW_STATE_FILE" | head -12)
    if echo "$PORT_BLOCK" | grep -q '"Port Power".*"enabled"'; then
        echo -e "${GREEN}✓ Port $port is enabled (was disabled)${NC}"
    else
        echo -e "${RED}✗ Port $port should be enabled but is still disabled${NC}"
        VERIFY_FAILED=1
    fi
done

if [ $VERIFY_FAILED -eq 1 ]; then
    echo -e "${RED}Verification failed!${NC}"
    echo "Continuing to restore original state..."
    echo ""
else
    echo -e "${GREEN}✓ All state changes verified successfully${NC}"
    echo ""
fi

# Step 4: Restore original state
echo -e "${YELLOW}[5/5] Restoring original state...${NC}"

# Enable ports that were originally ON
if [ ${#PORTS_ON_ARRAY[@]} -gt 0 ]; then
    echo "  Re-enabling ports: ${PORTS_ON_ARRAY[@]}"
    $BIN "$SWITCH" enable ${PORTS_ON_ARRAY[@]} > /dev/null
fi

# Disable ports that were originally OFF
if [ ${#PORTS_OFF_ARRAY[@]} -gt 0 ]; then
    echo "  Re-disabling ports: ${PORTS_OFF_ARRAY[@]}"
    $BIN "$SWITCH" disable ${PORTS_OFF_ARRAY[@]} > /dev/null
fi

echo -e "${GREEN}✓ Original state restored${NC}"
echo ""

# Final verification
echo -e "${YELLOW}Verifying restoration...${NC}"
sleep 2
$BIN "$SWITCH" settings > "$NEW_STATE_FILE"

RESTORE_FAILED=0
for port in ${PORTS_ON_ARRAY[@]:-}; do
    PORT_BLOCK=$(grep -A 10 "\"Port ID\".*\"$port\"" "$NEW_STATE_FILE" | head -12)
    if echo "$PORT_BLOCK" | grep -q '"Port Power".*"enabled"'; then
        echo -e "${GREEN}✓ Port $port restored to enabled${NC}"
    else
        echo -e "${RED}✗ Port $port should be enabled${NC}"
        RESTORE_FAILED=1
    fi
done

for port in ${PORTS_OFF_ARRAY[@]:-}; do
    PORT_BLOCK=$(grep -A 10 "\"Port ID\".*\"$port\"" "$NEW_STATE_FILE" | head -12)
    if echo "$PORT_BLOCK" | grep -q '"Port Power".*"disabled"'; then
        echo -e "${GREEN}✓ Port $port restored to disabled${NC}"
    else
        echo -e "${RED}✗ Port $port should be disabled${NC}"
        RESTORE_FAILED=1
    fi
done

echo ""
echo -e "${BLUE}========================================${NC}"
if [ $VERIFY_FAILED -eq 0 ] && [ $RESTORE_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests PASSED${NC}"
    echo -e "${BLUE}========================================${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests FAILED${NC}"
    echo -e "${BLUE}========================================${NC}"
    exit 1
fi
