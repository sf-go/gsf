package modules

import (
	"example/server/controllers"
	"example/server/models"
	"gsc/logger"
	"gsc/serialization"
	"gsf/peer"
	"gsf/service"
	"gsm/module"
)

type TestServerModule struct {
	*module.Module
}

func NewTestServerModule() *TestServerModule {
	return &TestServerModule{
		Module: module.NewModule(),
	}
}

func (testModule *TestServerModule) Initialize(service service.IService) {
	testModule.Module.Initialize(service)

	testModule.AddController(controllers.NewTestController())
	testModule.AddModel("TestModel", func() serialization.ISerializablePacket {
		return new(models.TestModel)
	})
	logger.Log.Debug("Initialize")
}

func (testModule *TestServerModule) InitializeFinish(service service.IService) {
	testModule.Module.InitializeFinish(service)

	logger.Log.Debug("InitializeFinish")
}

func (testModule *TestServerModule) Connected(peer peer.IPeer) {
	logger.Log.Debug("Connected")
}