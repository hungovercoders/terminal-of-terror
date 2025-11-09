#!/bin/bash
# filepath: test_lizard.sh

echo "🔍 Testing Lizard with Terminal of Terror project..."
echo "================================================"

echo "1. Testing basic lizard command:"
lizard --version

echo -e "\n2. Testing lizard on current directory:"
lizard -l go .

echo -e "\n3. Testing lizard with CSV output:"
lizard -l go --csv .

echo -e "\n4. Testing lizard with specific Go files:"
find . -name "*.go" -type f | head -5 | while read file; do
    echo "Analyzing: $file"
    lizard "$file"
done

echo -e "\n5. File structure:"
find . -name "*.go" -type f