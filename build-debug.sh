#!/bin/bash

# Debug script for Docker build issues
echo "🔍 Debugging Docker build for Golem Telegram Bot"
echo "================================================"

echo "📁 Current directory structure:"
ls -la

echo ""
echo "📦 Go module information:"
go mod graph

echo ""
echo "🔍 Checking pkg/golem directory:"
ls -la pkg/golem/

echo ""
echo "🔍 Checking examples directory:"
ls -la examples/

echo ""
echo "🔍 Testing local build:"
go build -o telegram-bot-test examples/telegram_bot.go
if [ $? -eq 0 ]; then
    echo "✅ Local build successful"
    rm -f telegram-bot-test
else
    echo "❌ Local build failed"
    exit 1
fi

echo ""
echo "🐳 Testing Docker build:"
docker build -t golem-telegram-bot-debug .
