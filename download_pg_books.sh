#!/bin/bash
# Downloads the first 100 fantasy books from Project Gutenberg using gutenberg CLI.

mkdir -p books

gutenberg search "subject:fantasy AND language:en" \
| jq -r '.key' | head -n 100 | while read -r KEY; do
    echo "Exporting book $KEY..."
    gutenberg text key:$KEY > my_books/pg-$KEY.txt
done
