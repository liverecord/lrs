package handlers

import "github.com/liverecord/lrs"

// @todo: needs a proper security there

// CallInit sends event for call to start
func (Ctx *ConnCtx) CallInit(frame lrs.Frame) {
	Ctx.Pool.Broadcast(frame, Ctx.Ws)
}

func (Ctx *ConnCtx) CallStop(frame lrs.Frame) {
	Ctx.Pool.Broadcast(frame, Ctx.Ws)
}

// CallCandidate sends ICECandidate data
func (Ctx *ConnCtx) CallCandidate(frame lrs.Frame) {
	Ctx.Pool.Broadcast(frame, Ctx.Ws)
}

// CallLocalDescription sends SDP
func (Ctx *ConnCtx) CallLocalDescription(frame lrs.Frame) {
	Ctx.Pool.Broadcast(frame, Ctx.Ws)
}