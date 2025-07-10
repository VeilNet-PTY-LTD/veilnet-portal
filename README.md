# VeilNet Portal

A self-hostable relay node that powers the decentralized backbone of the VeilNet network. The VeilNet Portal securely forwards encrypted traffic between VeilNet clients (called _Rifts_) without logging or revealing user identities.

By self-hosting a Portal, you help expand the VeilNet infrastructure, improve network resilience, and earn Measurement of Participation (MP) credits — all while retaining full control over your node.

## Features

- **Decentralized Relay**: Contribute bandwidth and relay capacity to the VeilNet network
- **Privacy-First**: No logging or tracking of user traffic or identities
- **Earn Rewards**: Receive MP credits for your contribution to network infrastructure
- **Full Control**: Maintain complete ownership and control of your relay node
- **Web Monitoring**: Real-time status interface for monitoring your portal
- **Cross-Platform**: Support for Linux, ARM64, and x86_64 architectures

## Prerequisites

Before setting up your VeilNet Portal, ensure you have:

- **Linux System**: Ubuntu 20.04+ or similar Linux distribution
- **Root Access**: Required for TUN device creation and network configuration
- **Network Connectivity**: Stable internet connection with open ports
- **Domain Name**: A domain name for your portal (optional but recommended)
- **Guardian Account**: Access to the VeilNet Guardian service

## Quick Start

### 1. Get Your Portal Token

1. Visit [guardian.veilnet.org](https://guardian.veilnet.org)
2. Sign in or create an account
3. Navigate to the Portal section
4. Generate a new Portal token
5. Note down your token, anchor name, and domain information

### 2. Choose Your Deployment Method

#### Option A: Docker (Recommended)

**Using Docker Compose:**

1. **Create docker-compose.yaml**:
```yaml
services:
  veilnet-portal:
    build: .
    container_name: veilnet-portal
    restart: unless-stopped
    privileged: true
    env_file:
      - .env
    ports:
      - "3000:3000"
    volumes:
      - /dev/net/tun:/dev/net/tun
```

2. **Create .env file**:
```bash
VEILNET_GUARDIAN_URL=https://guardian.veilnet.org
VEILNET_ANCHOR_TOKEN=your-portal-token-here
VEILNET_ANCHOR_NAME=your-anchor-name
VEILNET_DOMAIN_NAME=your-domain.com
VEILNET_REGION=us
VEILNET_PUBLIC=true
```

3. **Build and run**:
```bash
docker-compose up -d
```

**Using Docker directly:**
```bash
docker run -d \
  --name veilnet-portal \
  --privileged \
  --device=/dev/net/tun \
  -e VEILNET_GUARDIAN_URL=https://guardian.veilnet.org \
  -e VEILNET_ANCHOR_TOKEN=your-portal-token \
  -e VEILNET_ANCHOR_NAME=your-anchor-name \
  -e VEILNET_DOMAIN_NAME=your-domain.com \
  -e VEILNET_REGION=us \
  -e VEILNET_PUBLIC=true \
  -p 3000:3000 \
  veilnet/portal:latest
```

#### Option B: Native Installation

1. **Download the binary**:
Please download the binary from github repository releases.

2. **Make it executable**:
```bash
# For x86_64 Linux
chmod +x veilnet-portal

# For ARM64 Linux
chmod +x veilnet-portal-arm64
```

3. **Create configuration**:
```bash
# Create config.yaml
cat > config.yaml << EOF
guardian_url: "https://guardian.veilnet.org"
anchor_token: "your-portal-token-here"
anchor_name: "your-anchor-name"
domain_name: "your-domain.com"
region: "us"
public: true
EOF
```

4. **Run the portal**:
```bash
# For x86_64 Linux
sudo ./veilnet-portal

# For ARM64 Linux
sduo ./veilnet-portal-arm64
```

### 3. Verify Your Portal

1. **Check the web interface**: Visit `http://your-server:3000` to see your portal status
2. **Monitor logs**: Check the application logs for any errors
3. **Verify connectivity**: Ensure your portal is connected to the Guardian service

## Configuration

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `VEILNET_GUARDIAN_URL` | Guardian API endpoint | Yes | `https://guardian.veilnet.org` |
| `VEILNET_ANCHOR_TOKEN` | Your portal authentication token | Yes | - |
| `VEILNET_ANCHOR_NAME` | Your anchor name from Guardian | Yes | - |
| `VEILNET_DOMAIN_NAME` | Your domain name | Yes | - |
| `VEILNET_REGION` | Your portal region | Yes | - |
| `VEILNET_PUBLIC` | Whether to register as public portal | No | `true` |

### Configuration File

You can also use a `config.yaml` file:

```yaml
guardian_url: "https://guardian.veilnet.org"
anchor_token: "your-portal-token-here"
anchor_name: "your-anchor-name"
domain_name: "your-domain.com"
region: "us"
public: true
```

### Configuration Priority

Configuration values are loaded in this order (later overrides earlier):

1. **Default values** (hardcoded defaults)
2. **Config file values** (`config.yaml`)
3. **Environment variables** (with `VEILNET_` prefix)

## Web Interface

The portal provides a web interface at port 3000 that displays:

- **Anchor Information**: Name, domain, and region
- **Network Status**: Connection status and CIDR information

Access the interface at: `http://your-server:3000`

## Monitoring and Maintenance

### Logs

The portal uses structured logging. Check logs for detailed information:

```bash
# Docker logs
docker logs veilnet-portal -f

# System logs (if running as service)
sudo journalctl -u veilnet-portal -f

# Direct logs
sudo ./veilnet-portal 2>&1 | tee veilnet.log
```

### Health Checks

Monitor your portal's health:

```bash
# Check metrics endpoint
curl http://localhost:3000/metrics
```

### Updates

To update your portal:

```bash
# Docker
docker-compose pull
docker-compose up -d

# Native
# Download new binary and restart
```

## Troubleshooting

### Common Issues

**Permission Denied**
```bash
# Ensure running with sudo for native installation
sudo ./veilnet-portal

# For Docker, ensure --privileged flag is set
```

**TUN Device Creation Failed**
```bash
# Check if TUN module is loaded
lsmod | grep tun

# Load TUN module if needed
sudo modprobe tun

# For Docker, ensure --device=/dev/net/tun is set
```

**Network Configuration Failed**
```bash
# Check if iproute2 is installed
which ip

# Install if missing
sudo apt install iproute2
```

**Web Interface Not Accessible**
```bash
# Check if port 3000 is open
netstat -tlnp | grep 3000

# Check firewall rules
sudo iptables -L

# Ensure port 3000 is exposed in Docker
```

**Portal Not Connecting to Guardian**
```bash
# Check network connectivity
curl https://guardian.veilnet.org

# Verify token is correct
# Check logs for authentication errors
```

### Performance Optimization

- **Bandwidth**: Ensure sufficient upload/download bandwidth
- **CPU**: Portal is lightweight, minimal CPU requirements
- **Memory**: Typically uses 50-100MB RAM
- **Network**: Stable, low-latency connection recommended

## Security Considerations

- **Root Privileges**: Required for TUN device management
- **Token Security**: Keep your portal token secure and private
- **Network Isolation**: Consider running in isolated network environment
- **Updates**: Keep portal software updated for security patches

## Contributing to the Network

By running a VeilNet Portal, you're contributing to:

- **Network Resilience**: Providing relay capacity for the decentralized network
- **Privacy**: Enabling secure, private communication for users
- **Infrastructure**: Supporting the growth of the VeilNet ecosystem
- **Rewards**: Earning MP credits for your contribution

## Support

For help and support:

- **Documentation**: [www.veilnet.org/docs](https://www.veilnet.org/docs)
- **Community**: Join the VeilNet community discussions
- **Issues**: Report bugs and issues on GitHub
- **Guardian Support**: Contact support through the Guardian interface

## License

This project is licensed under the CC-BY-NC-ND-4.0 License.

## Changelog

### v1.0.0
- Initial release

### v1.0.1
- Fixed IP forwarding operation may accidentally overwrites host original settings when directly using binary

### v1.0.2
- Upgrade to Anchor Protocol version 2, enabled ML DSA signature

### v1.0.3
- Implemented TOFU in control plane