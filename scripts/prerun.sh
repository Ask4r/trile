#!/bin/sh

# Create `data` dir
mkdir -p data

# Create logs file
mkdir -p ~/.local/state/trile/logs
touch ~/.local/state/trile/logs/trile.log

# Kind reminder
echo "Don't forget about .env!"
