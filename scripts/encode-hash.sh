#!/bin/bash
# Script to encode Argon2 password hash to base64 for Helm --set usage
# This avoids Helm parsing issues with special characters (commas, $, etc.)
#
# Usage:
#   ./scripts/encode-hash.sh "$argon2i$v=19$m=4096,t=3,p=1$salt$hash"
#   Or pipe from generate-password-hash.go:
#   cd backend && go run ../scripts/generate-password-hash.go "mypassword" | xargs ../scripts/encode-hash.sh

if [ $# -eq 0 ]; then
    echo "Usage: $0 <argon2-hash>"
    echo ""
    echo "Example:"
    echo "  $0 '\$argon2i\$v=19\$m=4096,t=3,p=1\$salt\$hash'"
    echo ""
    echo "Or combine with generate-password-hash.go:"
    echo "  cd backend && go run ../scripts/generate-password-hash.go 'mypassword' | xargs ../scripts/encode-hash.sh"
    exit 1
fi

HASH="$1"

# Encode to base64
ENCODED=$(echo -n "$HASH" | base64 -w 0)

echo "Original hash:"
echo "$HASH"
echo ""
echo "Base64 encoded (use with --set admin.passwordHashBase64='...'):"
echo "$ENCODED"
echo ""
echo "Helm command example:"
echo "  helm install dkonsole ./helm/dkonsole --set admin.passwordHashBase64='$ENCODED'"
