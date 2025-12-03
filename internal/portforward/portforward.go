package portforward

import (
	"fmt"
	"io"
	"time"

	goios "github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/forward"

	"github.com/anthropics/maestro-ios-device/internal/utils"
)

type PortForwarder struct {
	entry      goios.DeviceEntry
	localPort  uint16
	devicePort uint16
	listener   interface{}
}

func New(entry goios.DeviceEntry, localPort, devicePort uint16) *PortForwarder {
	return &PortForwarder{
		entry:      entry,
		localPort:  localPort,
		devicePort: devicePort,
	}
}

func (p *PortForwarder) Start() error {
	listener, err := forward.Forward(p.entry, p.localPort, p.devicePort)
	if err != nil {
		return fmt.Errorf("port forward failed %d->%d: %w", p.localPort, p.devicePort, err)
	}
	p.listener = listener
	return nil
}

func (p *PortForwarder) Stop() {
	if p.listener == nil {
		return
	}
	if closer, ok := p.listener.(io.Closer); ok {
		closer.Close()
	}
	p.listener = nil
}

func (p *PortForwarder) Verify() error {
	time.Sleep(500 * time.Millisecond)
	if !utils.IsPortBusy(int(p.localPort)) {
		return fmt.Errorf("port %d not forwarded", p.localPort)
	}
	return nil
}
