package portal

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"os/exec"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/VeilNet-PTY-LTD/veilnet"
	"golang.zx2c4.com/wireguard/tun"
)

type Portal struct {
	anchor  *veilnet.Anchor
	device  tun.Device
	gateway string
	iface   string
	e       *echo.Echo
	once    sync.Once

	ctx    context.Context
	cancel context.CancelFunc
}

func NewPortal() *Portal {
	ctx, cancel := context.WithCancel(context.Background())

	return &Portal{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (p *Portal) Start(apiBaseURL, anchorToken, anchorName, domainName, region, hostInterface string, public bool) error {

	// Get the default gateway and interface
	gateway, iface, err := getDefaultGateway()
	if err != nil {
		return err
	}
	p.gateway = gateway
	p.iface = iface

	// Set bypass routes
	p.setBypassRoutes()

	// Create a new TUN device
	tun, err := tun.CreateTUN("veilnet", 1500)
	if err != nil {
		return err
	}
	p.device = tun

	// Create a new anchor
	a, err := veilnet.NewAnchor(true, public)
	if err != nil {
		return err
	}
	p.anchor = a

	// Start the anchor
	err = p.anchor.Start(apiBaseURL, anchorToken, anchorName, domainName, region)
	if err != nil {
		return err
	}

	// Get the IP address
	cidr, err := p.anchor.GetCIDR()
	if err != nil {
		return err
	}

	// Split CIDR into IP and netmask
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid CIDR format: %s", cidr)
	}

	ip := parts[0]
	netmask := parts[1]

	// Configure the TUN interface
	p.configTUN(ip, netmask)

	// Start the ingress and egress goroutines
	go p.ingress()
	go p.egress()

	// Create a new Echo app
	p.e = echo.New()
	p.e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseFS(templateFS, "template.html")),
	}
	p.e.GET("/", p.statusPage)
	p.e.GET("/metrics", echo.WrapHandler(p.anchor.Metrics.GetHandler()))

	go func() {
		veilnet.Logger.Sugar().Info("Starting Echo web server on :3000")
		if err := p.e.Start(":3000"); err != nil {
			veilnet.Logger.Sugar().Errorf("Failed to start Echo server: %v", err)
		}
	}()

	return nil
}

func (r *Portal) Stop() {
	r.once.Do(func() {
		// Stop the ingress and egress goroutines
		r.cancel()

		// Stop the anchor
		if r.anchor != nil {
			r.anchor.Stop()
		}

		// Remove the TUN interface firewall rules
		r.cleanUp()

		// Close the TUN device
		if r.device != nil {
			r.device.Close()
		}

		// Stop the Echo server
		if r.e != nil {
			r.e.Shutdown(context.Background())
		}
	})
}

func (p *Portal) Read(bufs [][]byte, batchSize int) (int, error) {
	return p.anchor.Read(bufs, batchSize)
}

func (p *Portal) Write(bufs [][]byte, sizes []int) (int, error) {
	return p.anchor.Write(bufs, sizes)
}

func (p *Portal) ingress() {
	bufs := make([][]byte, p.device.BatchSize())
	for {
		select {
		case <-p.anchor.Ctx.Done():
			veilnet.Logger.Sugar().Info("Portal ingress stopped")
			return
		default:
			n, err := p.anchor.Read(bufs, p.device.BatchSize())
			if err != nil {
				veilnet.Logger.Sugar().Errorf("failed to read from rift: %v", err)
				panic(err)
			}
			for i := 0; i < n; i++ {
				newBuf := make([]byte, 16+len(bufs[i]))
				copy(newBuf[16:], bufs[i])
				bufs[i] = newBuf
			}
			p.device.Write(bufs[:n], 16)
		}
	}
}

func (p *Portal) egress() {
	bufs := make([][]byte, p.device.BatchSize())
	sizes := make([]int, p.device.BatchSize())
	mtu, err := p.device.MTU()
	if err != nil {
		veilnet.Logger.Sugar().Errorf("failed to get TUN MTU: %v", err)
		// Use default MTU if we can't get the actual one
		mtu = 1500
	}
	// Pre-allocate buffers
	for i := range bufs {
		bufs[i] = make([]byte, mtu)
	}

	for {
		select {
		case <-p.anchor.Ctx.Done():
			veilnet.Logger.Sugar().Info("Portal egress stopped")
			return
		default:
			n, err := p.device.Read(bufs, sizes, 0)
			if err != nil {
				veilnet.Logger.Sugar().Errorf("failed to read from TUN: %v", err)
				continue
			}
			p.Write(bufs[:n], sizes[:n])
		}
	}
}

// getDefaultGatewayAndInterface returns the default gateway IP and the associated interface name on Linux
func getDefaultGateway() (gateway string, iface string, err error) {
	cmd := exec.Command("ip", "route", "show", "default")
	out, err := cmd.Output()
	if err != nil {
		veilnet.Logger.Sugar().Errorf("Failed to get default route: %v", err)
		return "", "", err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "default") {
			fields := strings.Fields(line)
			for i := 0; i < len(fields); i++ {
				if fields[i] == "via" && i+1 < len(fields) {
					gateway = fields[i+1]
				}
				if fields[i] == "dev" && i+1 < len(fields) {
					iface = fields[i+1]
				}
			}
			break
		}
	}

	if gateway == "" || iface == "" {
		err = fmt.Errorf("default gateway or interface not found")
		veilnet.Logger.Sugar().Errorf("Host default gateway or interface not found")
		return "", "", err
	}

	veilnet.Logger.Sugar().Infof("Found Host Default gateway: %s via interface %s", gateway, iface)
	return gateway, iface, nil
}

// setBypassRoutes adds routes to stun.cloudflare.com and turn.cloudflare.com via the specified gateway
func (p *Portal) setBypassRoutes() {
	hosts := []string{"stun.cloudflare.com", "turn.cloudflare.com", "guardian.veilnet.org"}

	for _, host := range hosts {
		ips, err := net.LookupIP(host)
		if err != nil {
			veilnet.Logger.Sugar().Errorf("Failed to resolve %s: %v", host, err)
			continue
		}

		for _, ip := range ips {
			if ip4 := ip.To4(); ip4 != nil {
				dest := ip4.String()

				// Add route
				cmd := exec.Command("ip", "route", "add", dest, "via", p.gateway, "dev", p.iface)
				cmd.Run()

			}
		}
	}
}

func (p *Portal) configTUN(ip, netmask string) error {

	// Flush existing IPs first
	cmd := exec.Command("ip", "addr", "flush", "dev", "veilnet")
	if err := cmd.Run(); err != nil {
		veilnet.Logger.Sugar().Errorf("failed to clear existing IPs: %v", err)
		return err
	}

	// Set the IP address
	cmd = exec.Command("ip", "addr", "add", fmt.Sprintf("%s/%s", ip, netmask), "dev", "veilnet")
	if err := cmd.Run(); err != nil {
		veilnet.Logger.Sugar().Errorf("failed to set IP address: %v", err)
		return err
	}
	veilnet.Logger.Sugar().Infof("VeilNet TUN IP address set to %s", ip)

	// Set iptables FORWARD
	cmd = exec.Command("iptables", "-A", "FORWARD", "-i", "veilnet", "-j", "ACCEPT")
	if err := cmd.Run(); err != nil {
		veilnet.Logger.Sugar().Errorf("failed to set inbound iptables FORWARD rules: %v", err)
		return err
	}
	cmd = exec.Command("iptables", "-A", "FORWARD", "-o", "veilnet", "-j", "ACCEPT")
	if err := cmd.Run(); err != nil {
		veilnet.Logger.Sugar().Errorf("failed to set outbound iptables FORWARD rules: %v", err)
		return err
	}
	veilnet.Logger.Sugar().Infof("Updated iptables FORWARD rules for VeilNet TUN")

	// Set up NAT
	cmd = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", p.iface, "-j", "MASQUERADE")
	if err := cmd.Run(); err != nil {
		veilnet.Logger.Sugar().Errorf("failed to set NAT rules: %v", err)
		return err
	}
	veilnet.Logger.Sugar().Infof("Set up NAT for VeilNet TUN")

	// Set the interface up
	cmd = exec.Command("ip", "link", "set", "up", "veilnet")
	if err := cmd.Run(); err != nil {
		veilnet.Logger.Sugar().Errorf("failed to set interface up: %v", err)
		return err
	}
	veilnet.Logger.Sugar().Infof("VeilNet TUN interface set to up")

	// Enable IP forwarding
	cmd = exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	if err := cmd.Run(); err != nil {
		veilnet.Logger.Sugar().Errorf("failed to enable IP forwarding: %v", err)
		return err
	}
	veilnet.Logger.Sugar().Infof("IP forwarding enabled")

	return nil
}

func (p *Portal) cleanUp() {

	// Remove iptables FORWARD rules
	cmd := exec.Command("iptables", "-D", "FORWARD", "-i", "veilnet", "-j", "ACCEPT")
	if err := cmd.Run(); err != nil {
		veilnet.Logger.Sugar().Warnf("failed to remove inbound iptables FORWARD rule: %v", err)
	}
	cmd = exec.Command("iptables", "-D", "FORWARD", "-o", "veilnet", "-j", "ACCEPT")
	if err := cmd.Run(); err != nil {
		veilnet.Logger.Sugar().Warnf("failed to remove outbound iptables FORWARD rule: %v", err)
	}

	// Remove NAT rule
	cmd = exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-o", p.iface, "-j", "MASQUERADE")
	if err := cmd.Run(); err != nil {
		veilnet.Logger.Sugar().Warnf("failed to remove NAT rule: %v", err)
	}

	// Disable IP forwarding
	cmd = exec.Command("sysctl", "-w", "net.ipv4.ip_forward=0")
	if err := cmd.Run(); err != nil {
		veilnet.Logger.Sugar().Warnf("failed to disable IP forwarding: %v", err)
	}
}
