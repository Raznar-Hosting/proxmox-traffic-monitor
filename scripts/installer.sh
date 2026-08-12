#!/bin/bash
set -e

# CHANGE THE URL AT ANYTIME.
BIN_URL="https://filebin.net/oa51ajrwqejuc9l7/pmx-tm_linux_amd64"
BIN_PATH="/usr/local/bin/pmx-tm"
SYMLINK="/usr/bin/pmx-tm"
SERVICE_PATH="/etc/systemd/system/pmx-tm.service"

echo "Downloading pmx-tm..."
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
systemctl enable pmx-tm
systemctl start pmx-tm

echo "Installation complete."
systemctl status pmx-tm --no-pager
