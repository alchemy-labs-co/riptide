#!/bin/bash

# Test script to verify reasoning tokens are captured

echo "Testing reasoning token capture..."
echo ""
echo "This test will send a simple math problem that should trigger reasoning."
echo "Look for a blue 'Reasoning:' label followed by the model's reasoning process."
echo ""
echo "To run this test:"
echo "1. Make sure DEEPSEEK_API_KEY is set"
echo "2. Run: ./riptide"
echo "3. Type: What is 25 * 37?"
echo "4. Watch for the reasoning tokens to appear"
echo ""
echo "Expected behavior:"
echo "- Blue dot (●) with 'Reasoning:' label"
echo "- Blue-colored reasoning text showing the calculation steps"
echo "- White dot (●) with 'Assistant>' label"
echo "- Final answer in white text"