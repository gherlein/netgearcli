#!/bin/bash

# POE Control Script
# Usage: ./poe-control.sh <switch-name> <enable|disable>
#
# This script controls all 8 POE ports on a Netgear switch using the poe-management tool.

set -e

# Check if correct number of arguments provided
if [ $# -ne 2 ]; then
    echo "Usage: $0 <switch-name> <enable|disable>"
    echo ""
    echo "Examples:"
    echo "  $0 tswitch1 enable"
    echo "  $0 tswitch2 disable"
    echo ""
    echo "This will control POE on ports 1-8 of the specified switch."
    exit 1
fi

SWITCH_NAME="$1"
ACTION="$2"

# Validate action parameter
case "$ACTION" in
    enable)
        POE_ACTION="enable"
        ;;
    disable)
        POE_ACTION="disable"
        ;;
    *)
        echo "Error: Action must be 'enable' or 'disable'"
        echo "You provided: '$ACTION'"
        exit 1
        ;;
esac

# Check if poe-management exists
POE_MANAGEMENT="$HOME/bin/poe-management"
if [ ! -x "$POE_MANAGEMENT" ]; then
    echo "Error: $POE_MANAGEMENT not found or not executable"
    echo "Please ensure the poe-management binary is installed in ~/bin/"
    exit 1
fi

echo "Controlling POE ports on switch: $SWITCH_NAME"
echo "Action: $ACTION"
echo "Ports: 1 2 3 4 5 6 7 8"
echo ""

# Execute the poe-management command
"$POE_MANAGEMENT" "$SWITCH_NAME" "$POE_ACTION" 1 2 3 4 5 6 7 8

echo ""
echo "POE control completed for switch $SWITCH_NAME"