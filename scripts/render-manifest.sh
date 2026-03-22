#!/usr/bin/env bash
set -euo pipefail

SOURCE="${1:-deploy/dkonsole.yaml}"
DOMAIN="${DKONSOLE_DOMAIN:-}"
INGRESS_CLASS="${DKONSOLE_INGRESS_CLASS:-nginx}"
ORIGIN_SCHEME="${DKONSOLE_ORIGIN_SCHEME:-https}"

read_source() {
    if [[ "$SOURCE" =~ ^https?:// ]]; then
        curl -fsSL "$SOURCE"
    else
        cat "$SOURCE"
    fi
}

render_base() {
    if [[ -n "$DOMAIN" ]]; then
        local allowed_origin="${ORIGIN_SCHEME}://${DOMAIN}"
        awk -v allowed_origin="$allowed_origin" '
            /# DKONSOLE_ALLOWED_ORIGINS_ENV/ {
                print "            - name: ALLOWED_ORIGINS"
                print "              value: \"" allowed_origin "\""
                next
            }
            { print }
        '
    else
        awk '!/# DKONSOLE_ALLOWED_ORIGINS_ENV/'
    fi
}

append_ingress() {
    [[ -n "$DOMAIN" ]] || return 0

    cat <<EOF
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: dkonsole
  namespace: dkonsole
  labels:
    app.kubernetes.io/name: dkonsole
    app.kubernetes.io/instance: dkonsole
spec:
  ingressClassName: ${INGRESS_CLASS}
  rules:
    - host: ${DOMAIN}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: dkonsole
                port:
                  number: 8080
EOF
}

read_source | render_base
append_ingress
