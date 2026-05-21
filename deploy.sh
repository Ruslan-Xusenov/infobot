#!/bin/bash

# Configuration
BRANCH="main"
SERVICE_NAME="infobot.service"
BINARY_NAME="bot_app"

# Navigate to the script's directory to ensure all git commands run in the correct context
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR" || { echo "Failed to navigate to project directory"; exit 1; }

echo "Checking for updates..."

# Fetch latest changes from remote repository
git fetch origin "$BRANCH" > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "Error: Failed to fetch from remote repository. Check network or Git credentials."
    exit 1
fi

# Get the commit hashes
LOCAL_HASH=$(git rev-parse HEAD)
REMOTE_HASH=$(git rev-parse origin/"$BRANCH")

if [ "$LOCAL_HASH" = "$REMOTE_HASH" ]; then
    echo "No updates available. System is up to date."
    exit 0
fi

echo "Updates found! Local: ${LOCAL_HASH::7}, Remote: ${REMOTE_HASH::7}"
echo "Updating code..."

# Discard local changes on the server to prevent merge conflicts
git reset --hard origin/"$BRANCH"

echo "Building Go application..."
# Compile Go application
go build -o "$BINARY_NAME" main.go
if [ $? -ne 0 ]; then
    echo "Error: Compilation failed!"
    exit 1
fi
echo "Go application compiled successfully."

echo "Restarting $SERVICE_NAME service..."
# Restart the systemd service
sudo systemctl restart "$SERVICE_NAME"
if [ $? -ne 0 ]; then
    echo "Error: Failed to restart systemd service '$SERVICE_NAME'."
    exit 1
fi

echo "Deployment completed successfully! Current version: ${REMOTE_HASH::7}"
exit 0
