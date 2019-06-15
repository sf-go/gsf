package gsm

import (
	"github.com/sf-go/gsf/src/gsc/crypto"
	"github.com/sf-go/gsf/src/gsc/logger"
	"github.com/sf-go/gsf/src/gsc/network"
	"github.com/sf-go/gsf/src/gsf/peer"
	"github.com/sf-go/gsf/src/gsf/service"
	"github.com/sf-go/gsf/src/gsm/module"
	"testing"
	"time"
)

type TestModule struct {
	*module.Module
}

func NewTestModule() *TestModule {
	return &TestModule{
		Module: module.NewModule(),
	}
}

func (testModule *TestModule) Initialize(service service.IService) {
	testModule.Module.Initialize(service)

	logger.Log.Debug("Initialize")
}

func (testModule *TestModule) Connected(peer peer.IPeer) {
	logger.Log.Debug("connected")
	peer.GetConnection().Close()
}

func (testModule *TestModule) Disconnected(peer peer.IPeer) {
	logger.Log.Debug("disconnected")
}

func (testModule *TestModule) InitializeFinish(service service.IService) {
	testModule.Module.InitializeFinish(service)

	logger.Log.Debug("InitializeFinish")
}

type Application struct {
}

func NewApplication() *Application {
	return &Application{}
}

func (application *Application) RegisterModule(moduleManager *module.ModuleManager) {
	moduleManager.AddModule("TestModule", NewTestModule())
}

func (application *Application) SetLogConfig(config *logger.LogConfig) {

}

func (application *Application) SetNetConfig(config *network.NetConfig) {
	config.BufferSize = 50
	config.Address = "127.0.0.1"
	config.Port = 8889
	config.ConnectTimeout = 3
}

func (application *Application) SetCryptoConfig(config *crypto.CryptoConfig) {

}

func TestRunServer(t *testing.T) {
	serverApplication := NewApplication()
	RunServer(serverApplication, nil)

	clientApplication := NewApplication()
	RunClient(clientApplication, nil)

	time.Sleep(3 * time.Second)
}
