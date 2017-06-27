// +build darwin dragonfly freebsd linux netbsd openbsd solaris windows

package executor

import (
	"fmt"
	"io"

	syslog "github.com/RackSec/srslog"

	"github.com/hashicorp/nomad/client/driver/logging"
)

func (e *UniversalExecutor) LaunchSyslogServer() (*SyslogServerState, error) {
	// Ensure the context has been set first
	if e.ctx == nil {
		return nil, fmt.Errorf("SetContext must be called before launching the Syslog Server")
	}

	e.syslogChan = make(chan *logging.SyslogMessage, 2048)
	l, err := e.getListener(e.ctx.PortLowerBound, e.ctx.PortUpperBound)
	if err != nil {
		return nil, err
	}
	e.logger.Printf("[DEBUG] syslog-server: launching syslog server on addr: %v", l.Addr().String())
	if err := e.configureLoggers(); err != nil {
		return nil, err
	}

	e.syslogServer = logging.NewSyslogServer(l, e.syslogChan, e.logger)
	go e.syslogServer.Start()
	go e.collectLogs(e.lre, e.lro)
	syslogAddr := fmt.Sprintf("%s://%s", l.Addr().Network(), l.Addr().String())
	return &SyslogServerState{Addr: syslogAddr}, nil
}

func (e *UniversalExecutor) collectLogs(we io.Writer, wo io.Writer) {
	for logParts := range e.syslogChan {
		// If the severity of the log line is err then we write to stderr
		// otherwise all messages go to stdout
		if logParts.Severity&syslog.LOG_ERR == 1 {
			e.lre.Write(logParts.Message)
			e.lre.Write([]byte{'\n'})
		} else {
			e.lro.Write(logParts.Message)
			e.lro.Write([]byte{'\n'})
		}
	}
}
