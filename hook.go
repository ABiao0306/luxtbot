package luxtbot

// When a bot instance onconnect to a CQ server,
// all of the OnconnectHook will be called.
// If the hook return an error which is 
// not nil, this connect will be closed, 
// and then, the Disconnecthooks will be called.
type OnconnectHook func(bInfo BotInfo) error

func (hook OnconnectHook) AddToHookChain() {
    onConnectChain = append(onConnectChain, hook)
}

func MakeOnconnectHook(task func(bInfo BotInfo) error) OnconnectHook{
    hook := task
    return hook
}

// When a connect is closed, this hook will be called.
type DisconnectHook func(bInfo BotInfo)

func (hook DisconnectHook) AddToHookChain() {
    disConnectChain = append(disConnectChain, hook)
}

func MakeDisconnectHook(task func(bInfo BotInfo)) DisconnectHook{
    hook := task
    return hook
}

// When an event is received from CQ server,
// all of the EventInHook will be called.
// If the hook return a not nil error,
// the event will be aborted.
type EventInHook func(e *Event, bInfo BotInfo) error

func (hook EventInHook) AddToHookChain() {
    eventInChain = append(eventInChain, hook)
}

func MakeEventInHook(task func(e *Event, bInfo BotInfo) error) EventInHook{
    hook := task
    return hook
}

// Before send an ApiPost to the CQ server,
// all of the  BeforeApiOutHook will be called.
// if the hook return a not nil error,
// this ApiPost will be aborted
type BeforeApiOutHook func(apiPost *ApiPost ,bInfo BotInfo) error

func (hook BeforeApiOutHook) AddToHookChain() {
    beforeApiOutChain = append(beforeApiOutChain, hook)
}

func MakeBeforeApiOutHook(task func(apiPost *ApiPost, bInfo BotInfo) error) BeforeApiOutHook{
    hook := task
    return hook
}

