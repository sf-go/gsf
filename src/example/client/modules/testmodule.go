package modules

import (
	"example/client/controllers"
	"example/client/models"
	"gsc/logger"
	"gsc/serialization"
	"gsf/peer"
	"gsf/service"
	"gsm/module"
	"strconv"
)

type TestClientModule struct {
	*module.Module
}

func NewTestClientModule() *TestClientModule {
	return &TestClientModule{
		Module: module.NewModule(),
	}
}

func (testModule *TestClientModule) Initialize(service service.IService) {
	testModule.Module.Initialize(service)

	testModule.AddController(controllers.NewTestController())
	testModule.AddModel("TestModel", func() serialization.ISerializablePacket {
		return new(models.TestModel)
	})
	logger.Log.Debug("Initialize")
}

func (testModule *TestClientModule) Connected(peer peer.IPeer) {
	controller := controllers.NewTestController()
	result := controller.Invoke("Test", peer, func() []interface{} {
		return []interface{}{
			10000,
			&models.TestModel{
				Name: "wwj",
				Age:  500,
			},
		}
	})

	logger.Log.Debug(strconv.Itoa(result[0].(int)))

	controller.AsyncInvoke("Test", peer, func() []interface{} {
		return []interface{}{
			10000,
			&models.TestModel{
				Name: "wwj",
				Age:  500,
			},
		}
	}, func(result []interface{}) {
		logger.Log.Debug(strconv.Itoa(result[0].(int)))
	})
}

func (testModule *TestClientModule) InitializeFinish(service service.IService) {
	testModule.Module.InitializeFinish(service)

	logger.Log.Debug("InitializeFinish")
}