#!/bin/bash

echo "{state=1}"
echo "{listen=192.168.1.100}"
echo "Checking network..."
IP=$(ip route get 1 2>/dev/null | awk '{print $NF;exit}')
echo "Your IP Address is\n$IP"

exit 0