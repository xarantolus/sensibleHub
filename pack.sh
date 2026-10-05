#!/bin/bash
set -euo pipefail

rm sensibleHub.zip || true
rm sensibleHub || true
rm sensibleHub.exe || true

# Use first argument for name if possible, fallback to sensibleHub.zip
NAME=${1:-sensibleHub.zip}

if [ ! -f frontend/dist/index.html ]; then
	make frontend-build
fi

make server BINARY="sensibleHub$(go env GOEXE)"

zip -r "$NAME" LICENSE README.md config.json sensibleHub*

echo "You can now extract $NAME somewhere on the target system and run the executable"
