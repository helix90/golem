#!/bin/bash

echo "=============================================="
echo "GOLEM WEATHER CLI TEST"
echo "=============================================="
echo ""
echo "This script tests the weather SRAIX functionality"
echo "Please run this and share the output with Claude"
echo ""

# Set fake API key (or use real one if you have it)
export PIRATE_WEATHER_API_KEY="${PIRATE_WEATHER_API_KEY:-fake-demo-key}"

# Check if binary exists
if [ ! -f "./build/golem" ]; then
    echo "ERROR: ./build/golem not found"
    echo "Building now..."
    go build -o build/golem ./cmd/golem
    chmod +x build/golem
fi

echo "Binary: $(file build/golem | head -c 100)"
echo "API Key: ${PIRATE_WEATHER_API_KEY}..."
echo ""

# Create test commands
cat > /tmp/weather_test_commands.txt << 'COMMANDS'
load testdata
session create weather-test
chat my location is honolulu
chat what is my location
chat what is the weather
quit
COMMANDS

echo "Commands to execute:"
cat /tmp/weather_test_commands.txt | sed 's/^/  > /'
echo ""
echo "=============================================="
echo "OUTPUT:"
echo "=============================================="

# Run test
./build/golem interactive < /tmp/weather_test_commands.txt 2>&1 | tee /tmp/weather_test_output.txt

echo ""
echo "=============================================="
echo "ANALYSIS:"
echo "=============================================="

# Check for issues
if grep -q "<think>" /tmp/weather_test_output.txt; then
    echo "❌ ISSUE: <think> tags visible in output"
fi

if grep -q "Unable to find that location" /tmp/weather_test_output.txt; then
    echo "❌ ISSUE: Geocoding failed"
fi

if grep -q "Virtual" /tmp/weather_test_output.txt; then
    echo "❌ ISSUE: Location showing as 'Virtual'"
fi

if grep -q "coordinates: [0-9]" /tmp/weather_test_output.txt; then
    echo "✅ PASS: Coordinates found"
fi

if grep -q "honolulu" /tmp/weather_test_output.txt; then
    echo "✅ PASS: Location honolulu mentioned"
fi

echo ""
echo "Full output saved to: /tmp/weather_test_output.txt"
echo ""
echo "Please share this output with Claude"
