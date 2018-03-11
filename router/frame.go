package router

import (
	"errors"

	"github.com/liverecord/server/common/frame"
	"github.com/liverecord/server/handlers"
)

// ErrNotRegisteredFrameType shows that we tried to process frame with type we had not registered
var ErrNotRegisteredFrameType = errors.New("not registered frame type")

// FrameHandlerFunc is a type of function for processing frame
type FrameHandlerFunc func(ctx *handlers.AppContext, f frame.Frame) (frame.Frame, error)

// Frame is a frames router
type Frame struct {
	handlers map[string]FrameHandlerFunc
}

// NewFrame func creates frame router
func NewFrame() *Frame {
	return &Frame{
		handlers: make(map[string]FrameHandlerFunc),
	}
}

// AddHandler func registers new handler for frame type
func (f *Frame) AddHandler(frameType string, handler FrameHandlerFunc) {
	f.handlers[frameType] = handler
}

// Process func pass data to handler
func (f *Frame) Process(ctx *handlers.AppContext, input frame.Frame) (frame.Frame, error) {
	handler, ok := f.handlers[input.Type]
	if !ok {
		return frame.Frame{}, ErrNotRegisteredFrameType
	}

	return handler(ctx, input)
}
