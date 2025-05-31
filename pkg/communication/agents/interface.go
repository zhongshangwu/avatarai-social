package agents

import "context"

type Runner interface {
	GetName() string
	GetDescription() string
	Ctrl(ctx *InvokeContext, ctrlType CtrlType)
	Invoke(ctx *InvokeContext) error
}

type InvokeContext struct {
	Context context.Context
}

func (ic *InvokeContext) isInvokeContext() {}
