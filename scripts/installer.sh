#!/bin/bash
set -e

# CHANGE THE URL AT ANYTIME.
BIN_URL="https://filebin.net/oa51ajrwqejuc9l7/proxmox-traffic-monitor_linux_amd64"
BIN_PATH="/usr/local/bin/proxmox-traffic-monitor"
SYMLINK="/usr/bin/proxmox-traffic-monitor"
SERVICE_PATH="/etc/systemd/system/proxmox-traffic-monitor.service"

echo "Downloading proxmox-traffic-monitor..."
curl -fsSL -o "$BIN_PATH" "$BIN_URL"

echo "Setting executable permission..."
chmod +x "$BIN_PATH"

echo "Creating symlink..."
ln -sf "$BIN_PATH" "$SYMLINK"

echo "Creating systemd service..."
cat > "$SERVICE_PATH" <<EOF
[Unit]
Description=Proxmox Traffic Monitor
After=network.target
Wants=network.target

[Service]
Type=simple
ExecStart=$BIN_PATH
Restart=always
RestartSec=5
User=root
WorkingDirectory=/root

[Install]
WantedBy=multi-user.target
EOF

echo "Reloading systemd..."
systemctl daemon-reexec
systemctl daemon-reload

echo "Enabling and starting service..."
systemctl enable proxmox-traffic-monitor
systemctl start proxmox-traffic-monitor

echo "Installation complete."
systemctl status proxmox-traffic-monitor --no-pager
