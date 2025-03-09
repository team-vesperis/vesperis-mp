package register

import (
	"errors"
	"net"

	"fmt"

	"github.com/team-vesperis/vesperis-mp/mp/mp/datasync"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

var (
	p          *proxy.Proxy
	logger     *zap.SugaredLogger
	proxy_name string
)

func InitializeRegister(proxy *proxy.Proxy, log *zap.SugaredLogger, pn string) {
	p = proxy
	logger = log
	proxy_name = pn
	logger.Info("Successfully initialized register.")
}

func RegisterServer(server_name string, host string, port int) error {
	address := host + ":" + fmt.Sprint(port)

	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		logger.Error("Error resolving address: ", err)
		return err
	}
	server_info := proxy.NewServerInfo(server_name, tcpAddr)
	_, err = p.Register(server_info)

	if err != nil {
		logger.Error("Error registering server: ", err)
		return err
	}

	err = datasync.RegisterServer(proxy_name, server_name)
	if err != nil {
		logger.Error("Error registering server in the database: ", err)
	}

	logger.Info("Successfully registered a new server called: ", server_name, " with address: ", address)
	return nil
}

func UnregisterServer(server_name string) error {
	server := p.Server(server_name)

	found := p.Unregister(server.ServerInfo())
	if !found {
		err := errors.New("server not found to unregister")
		logger.Error(err)
		return err
	}

	err := datasync.UnregisterServer(proxy_name, server_name)
	if err != nil {
		logger.Error("Error unregistering server in the database: ", err)
	}

	logger.Info("Successfully unregistered a server called: ", server_name, " with address: ", server.ServerInfo().Addr())
	return nil
}
