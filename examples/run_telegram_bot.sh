#!/bin/bash

# Golem Telegram Bot Runner Script
# This script helps you set up and run the Golem Telegram Bot

set -e

echo "🤖 Golem Telegram Bot Setup"
echo "=========================="
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.19+ first."
    echo "   Visit: https://golang.org/doc/install"
    exit 1
fi

echo "✅ Go is installed: $(go version)"
echo

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -d "pkg/golem" ]; then
    echo "❌ Please run this script from the Golem project root directory"
    echo "   (where go.mod and pkg/golem/ are located)"
    exit 1
fi

echo "✅ Running from Golem project directory"
echo

# Check for required environment variables
if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo "❌ TELEGRAM_BOT_TOKEN environment variable is not set"
    echo
    echo "To get a bot token:"
    echo "1. Message @BotFather on Telegram"
    echo "2. Use /newbot command"
    echo "3. Follow the instructions"
    echo "4. Set the token: export TELEGRAM_BOT_TOKEN='your_token_here'"
    echo
    exit 1
fi

echo "✅ Telegram bot token is set"
echo

# Set default AIML path if not provided
if [ -z "$AIML_PATH" ]; then
    if [ -d "testdata" ]; then
        export AIML_PATH="testdata"
        echo "✅ Using default AIML path: testdata"
    else
        echo "❌ AIML_PATH environment variable is not set and testdata directory not found"
        echo
        echo "Please set AIML_PATH to your AIML files directory:"
        echo "export AIML_PATH='/path/to/your/aiml/files'"
        echo
        exit 1
    fi
else
    echo "✅ Using AIML path: $AIML_PATH"
fi

# Check if AIML path exists
if [ ! -d "$AIML_PATH" ]; then
    echo "❌ AIML path does not exist: $AIML_PATH"
    exit 1
fi

echo "✅ AIML path exists and is accessible"
echo

# Install dependencies
echo "📦 Installing dependencies..."
go mod tidy
go get github.com/go-telegram/bot

echo "✅ Dependencies installed"
echo

# Check if verbose mode is enabled
if [ "$VERBOSE" = "true" ]; then
    echo "🔧 Verbose mode enabled"
else
    echo "🔧 Verbose mode disabled (set VERBOSE=true to enable)"
fi

echo

# Display configuration
echo "📋 Bot Configuration:"
echo "   Token: ${TELEGRAM_BOT_TOKEN:0:10}..."
echo "   AIML Path: $AIML_PATH"
echo "   Verbose: ${VERBOSE:-false}"
echo

# Ask for confirmation
read -p "🚀 Start the Telegram bot? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Bot startup cancelled"
    exit 0
fi

echo
echo "🚀 Starting Golem Telegram Bot..."
echo "   Press Ctrl+C to stop the bot"
echo

# Run the bot
go run examples/telegram_bot.go
