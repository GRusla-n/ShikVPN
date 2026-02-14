#!/bin/bash
set -e

echo "=== ShikVPN Integration Test ==="

# Step 1: Wait for server API to be ready
echo "[1/4] Waiting for server API..."
for i in $(seq 1 30); do
    if curl -s -o /dev/null -w "%{http_code}" http://172.20.0.10:8080/api/v1/register 2>/dev/null | grep -q "4[0-9][0-9]\|200"; then
        echo "  Server API is up"
        break
    fi
    if [ "$i" -eq 30 ]; then
        echo "  FAIL: Server API did not come up in 30s"
        exit 1
    fi
    sleep 1
done

# Step 2: Start VPN client in background
echo "[2/4] Starting VPN client..."
./vpn-client -config /app/configs/client.toml &
CLIENT_PID=$!
sleep 5

# Check client is still running
if ! kill -0 $CLIENT_PID 2>/dev/null; then
    echo "  FAIL: VPN client exited unexpectedly"
    wait $CLIENT_PID || true
    exit 1
fi
echo "  Client started (PID $CLIENT_PID)"

# Step 3: Verify VPN tunnel by pinging server's VPN IP
echo "[3/4] Pinging server VPN IP (10.0.0.1)..."
if ping -c 3 -W 5 10.0.0.1; then
    echo "  PASS: VPN tunnel is working!"
else
    echo "  FAIL: Cannot reach server through VPN tunnel"
    kill $CLIENT_PID 2>/dev/null || true
    exit 1
fi

# Step 4: Verify we got an IP in the VPN subnet
echo "[4/4] Checking assigned VPN interface..."
ip addr show | grep "10.0.0" && echo "  PASS: VPN address assigned"

# Cleanup
echo ""
echo "=== All tests PASSED ==="
kill $CLIENT_PID 2>/dev/null || true
wait $CLIENT_PID 2>/dev/null || true
exit 0
