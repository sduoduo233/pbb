#!/bin/bash

set -e

if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run as root."
    exit 1
fi

agnet_url="$1"
agent_secret="$2"
if [ -z "$agnet_url" ] || [ -z "$agent_secret" ]; then
    echo "Usage: $0 <agent_url> <agent_secret>"
    exit 1
fi

uname=$(uname -m)
case $uname in
    x86_64)
        arch="amd64"
        ;;
    aarch64|arm64)
        arch="arm64"
        ;;
    i386|i686)
        arch="386"
        ;;
    *arm*)
        arch="arm32-v7a"
        ;;
    *)
        echo "Unsupported architecture: $uname"
        exit 1
        ;;
esac

mkdir -p /opt/pbb
binary_url="https://dl.exec.li/agent-$arch"
echo "> Downloading $binary_url to /opt/pbb/agent"
curl -L "$binary_url" > /opt/pbb/agent

echo "> Setting permissions for /opt/pbb/agent"
chmod +x /opt/pbb/agent

echo "> Writing environment variables to /opt/pbb/.env"
echo "AGENT_URL=$agnet_url" > /opt/pbb/.env
echo "AGENT_SECRET=$agent_secret" >> /opt/pbb/.env
chmod 600 /opt/pbb/.env

is_systemd=$(command -v systemctl || true)
is_openrc=$(command -v rc-status || true)

if [ "$is_systemd" ]; then

    echo "> Creating systemd service file at /etc/systemd/system/pbb-agent.service"
    cat <<EOF > /etc/systemd/system/pbb-agent.service
[Unit]
Description=PBB Agent Service
After=network.target

[Service]
Type=simple
ExecStart=/opt/pbb/agent
Restart=on-failure
User=root
WorkingDirectory=/opt/pbb

[Install]
WantedBy=multi-user.target
EOF

    echo "> Starting systemd service"
    systemctl daemon-reload
    systemctl enable pbb-agent.service
    systemctl restart pbb-agent.service

elif [ "$is_openrc" ]; then

    echo "> Creating OpenRC service file at /etc/init.d/pbb-agent"
    cat <<EOF > /etc/init.d/pbb-agent
#!/sbin/openrc-run
description="PBB Agent Service"
command="/opt/pbb/agent"
directory="/opt/pbb"
depend() {
    need net
}

EOF

    echo "> Starting OpenRC service"

    chmod +x /etc/init.d/pbb-agent
    rc-update add pbb-agent default

    /etc/init.d/pbb-agent restart

else
    echo "Neither systemd nor OpenRC found. Please create a service file manually."
fi

echo "> PBB agent installation complete."