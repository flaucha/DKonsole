module github.com/example/k8s-view

go 1.24.0

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/gorilla/websocket v1.5.1
	golang.org/x/crypto v0.26.0
	k8s.io/apimachinery v0.29.0
	k8s.io/client-go v0.29.0
	k8s.io/metrics v0.29.0
	sigs.k8s.io/yaml v1.4.0
)

require golang.org/x/time v0.14.0 // indirect
