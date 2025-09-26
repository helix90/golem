#!/bin/bash

echo "🐳 Testing Docker build with replace directive..."
echo "=============================================="

# Test the minimal Dockerfile first
echo "Testing minimal Dockerfile..."
docker build -f Dockerfile.minimal -t golem-test-minimal . --no-cache

if [ $? -eq 0 ]; then
    echo "✅ Minimal Dockerfile build successful!"
    echo "Testing main Dockerfile..."
    docker build -t golem-test-main . --no-cache
    
    if [ $? -eq 0 ]; then
        echo "✅ Main Dockerfile build successful!"
        echo "🎉 All Docker builds working correctly!"
    else
        echo "❌ Main Dockerfile build failed"
        exit 1
    fi
else
    echo "❌ Minimal Dockerfile build failed"
    exit 1
fi
