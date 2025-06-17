#!/bin/bash
# Deploy script for Raspberry Pi

set -e

# Configuration
PI_HOST="${PI_HOST:-pi@raspberrypi.local}"
PI_DIR="/home/pi/dnd-bot"

echo "🚀 Deploying D&D Bot to Raspberry Pi..."

# Build for ARM architecture (Pi)
echo "📦 Building Docker image for ARM..."
docker buildx build --platform linux/arm64,linux/arm/v7 -t dnd-bot:latest .

# Save the image
echo "💾 Saving Docker image..."
docker save dnd-bot:latest | gzip > dnd-bot.tar.gz

# Copy files to Pi
echo "📤 Copying files to Pi..."
scp dnd-bot.tar.gz docker-compose.yml .env ${PI_HOST}:${PI_DIR}/

# Deploy on Pi
echo "🎯 Deploying on Pi..."
ssh ${PI_HOST} << 'EOF'
cd /home/pi/dnd-bot
echo "Loading Docker image..."
docker load < dnd-bot.tar.gz
echo "Starting services..."
docker-compose --profile production up -d
echo "Cleaning up..."
rm dnd-bot.tar.gz
docker-compose ps
EOF

echo "✅ Deployment complete!"
echo "📊 Redis Commander available at: http://${PI_HOST}:8081"