package register

import (
	"errors"
	"net"

	"fmt"

	"github.com/team-vesperis/vesperis-mp/mp/share"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

var (
	p      *proxy.Proxy
	logger *zap.SugaredLogger
)

func InitializeRegister(proxy *proxy.Proxy, log *zap.SugaredLogger) {
	p = proxy
	logger = log
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

	share.RegisterServer(server_name)
	logger.Info("Successfully registered a new server called: ", server_name, " with address: ", address)
	return nil
}

func UnregisterServer(server_name string) error {
	server := p.Server(server_name)
	if server == nil {
		err := errors.New("server not found to unregister")
		logger.Error(err)
		return err
	}

	p.Unregister(server.ServerInfo())
	share.UnregisterServer(server_name)
	logger.Info("Successfully unregistered a server called: ", server_name, " with address: ", server.ServerInfo().Addr())
	return nil
}
