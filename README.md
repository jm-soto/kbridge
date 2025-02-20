# kbridge

A simple Kubernetes port-forwarding tool that really simplifies connecting to internal kubernetes services

## Installation

```bash
go install github.com/kbridge/kbridge@latest
```

## Usage

kbridge simplifies the process of creating port-forwards to Kubernetes services. It offers two modes:

1. **Interactive Mode**: If you don't specify a service name, kbridge will:
   - List all available services in the namespace
   - Let you select a service from the list
   - If the service has multiple ports, let you choose one

2. **Direct Mode**: If you know the service name, you can connect directly:
   - Automatically selects a random local port
   - If the service has multiple ports, lets you choose one
   - Shows a clear local URL for connecting

### Service Port-Forward

```bash
# Forward to a service
kbridge service [service-name] -n [namespace]

# Interactive mode (no service name)
kbridge service -n [namespace]
```

### Options

- `-n, --namespace`: Kubernetes namespace (default: "default")
- `-L, --local-port`: Local port to forward from (random if not specified)
- `-p, --port`: Remote port to forward to (service port if not specified)

## License

MIT License
