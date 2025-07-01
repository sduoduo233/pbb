#!/bin/bash

set -e

if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run as root."
    exit 1
fi

echo "> Stopping existing service"
systemctl stop pbb-hub || true
/etc/init.d/pbb-hub stop || true

uname=$(uname -m)
case $uname in
    x86_64)
        arch="amd64"
        ;;
    aarch64|arm64)
        arch="arm64"
        ;;
    *)
        echo "Unsupported architecture: $uname"
        exit 1
        ;;
esac

mkdir -p /opt/pbb
binary_url="https://dl.exec.li/hub-$arch"
echo "> Downloading $binary_url to /opt/pbb/hub"
curl -L "$binary_url" > /opt/pbb/hub

echo "> Setting permissions for /opt/pbb/hub"
chmod +x /opt/pbb/hub

is_systemd=$(command -v systemctl || true)
is_openrc=$(command -v rc-status || true)

if [ "$is_systemd" ]; then

    echo "> Creating systemd service file at /etc/systemd/system/pbb-hub.service"
    cat <<EOF > /etc/systemd/system/pbb-hub.service
[Unit]
Description=PBB Hub Service
After=network.target

[Service]
Type=simple
ExecStart=/opt/pbb/hub
Restart=on-failure
User=root
WorkingDirectory=/opt/pbb

[Install]
WantedBy=multi-user.target
EOF

    echo "> Starting systemd service"
    systemctl daemon-reload
    systemctl enable pbb-hub.service
    systemctl restart pbb-hub.service

elif [ "$is_openrc" ]; then

    echo "> Creating OpenRC service file at /etc/init.d/pbb-hub"
    cat <<EOF > /etc/init.d/pbb-hub
#!/sbin/openrc-run
description="PBB Hub Service"
command="/opt/pbb/hub"
directory="/opt/pbb"
depend() {
    need net
}

EOF

    echo "> Starting OpenRC service"

    chmod +x /etc/init.d/pbb-hub
    rc-update add pbb-hub default

    /etc/init.d/pbb-hub restart

else
    echo "Neither systemd nor OpenRC found. Please create a service file manually."
fi

echo "> PBB Hub installation complete."