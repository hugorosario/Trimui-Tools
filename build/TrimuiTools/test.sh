#!/bin/bash

echo "Checking network..."
IP=$(ip route get 1 2>/dev/null | awk '{print $NF;exit}')
echo "Your IP Address is\n$IP\nSee how cool is that?\nGreat success!"

exit 0