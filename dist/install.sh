#!/bin/bash

# Ensure the script is run with root privileges
if [ "$EUID" -ne 0 ]; then
  echo "❌ Please run this installer as root (e.g., sudo ./install.sh)"
  exit 1
fi

echo "🛡️ Step 1: Creating dedicated 'nvr' service user..."
# Only create the user if it doesn't already exist
if ! id -u nvr > /dev/null 2>&1; then
    useradd -r -s /bin/false nvr
    echo "   User 'nvr' created successfully."
else
    echo "   User 'nvr' already exists, skipping."
fi

echo "📦 Step 2: Installing FFmpeg runtime dependencies..."
apt-get update
apt-get install -y ffmpeg libavformat58 libavcodec58 libavutil56 libswscale5 ca-certificates

echo "📂 Step 3: Creating secure system directories..."
mkdir -p /opt/nvr/bin
mkdir -p /opt/nvr/recordings
mkdir -p /opt/nvr/db
mkdir -p /etc/nvr

echo "⚙️ Step 4: Installing binaries and configuration..."
cp nvr_service /opt/nvr/bin/
cp nvr_worker /opt/nvr/bin/
cp -r web /opt/nvr/bin/web
# Assumes config.json is bundled in your tarball next to the installer
cp config.json /opt/nvr/bin/config.json

echo "🔒 Step 5: Locking down permissions..."
# Give ownership to the new nvr user
chown -R nvr:nvr /opt/nvr
chown -R nvr:nvr /etc/nvr

# Set strict directory permissions (Owner: Read/Write/Execute, Group: Read/Execute, Other: None)
chmod 750 /opt/nvr/bin
chmod 755 /opt/nvr/recordings
chmod 750 /opt/nvr/db
chmod 750 /etc/nvr

# Ensure binaries are executable and config is readable by the nvr group
chmod +x /opt/nvr/bin/nvr_service
chmod +x /opt/nvr/bin/nvr_worker
chmod 640 /opt/nvr/bin/config.json

echo "🔄 Step 6: Setting up systemd background service..."
cp nvr.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable nvr.service
systemctl restart nvr.service

echo "✅ Secure NVR installation complete!"
echo "Use 'systemctl status nvr.service' to check its status."