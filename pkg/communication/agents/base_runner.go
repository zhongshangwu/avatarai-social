package agents

import (
	"github.com/sirupsen/logrus"
)

type BaseRunner struct {
	name        string
	description string
	ctrl        chan CtrlType
}

func NewBaseRunner(name, description string) *BaseRunner {
	return &BaseRunner{
		name:        name,
		description: description,
	}
}

func (a *BaseRunner) GetName() string {
	return a.name
}

func (a *BaseRunner) GetDescription() string {
	return a.description
}

func (a *BaseRunner) Ctrl(ctx *InvokeContext, ctrlType CtrlType) {
	a.ctrl <- ctrlType
}

func (a *BaseRunner) Invoke(ctx InvokeContext) error {
	logrus.Warnf("BaseRunner.Invoke 被调用，但没有具体实现")
	return nil
}
