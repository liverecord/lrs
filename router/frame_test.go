package router

import (
	"testing"

	"github.com/liverecord/server/common/frame"
	"github.com/liverecord/server/handlers"
)

func TestFrame_AddHandler(t *testing.T) {
	handler := func(ctx *handlers.AppContext, f frame.Frame) (frame.Frame, error) {
		return frame.Frame{}, nil
	}

	router := NewFrame()
	router.AddHandler("test", handler)

	_, ok := router.handlers["test"]
	if !ok {
		t.Error("I expected to see handler for \"test\" but do not see it")
	}
}

func TestFrame_ProcessNotExistsHandler(t *testing.T) {
	router := NewFrame()
	ctx := &handlers.AppContext{}
	f := frame.Frame{Type: "test"}

	_, err := router.Process(ctx, f)
	if err != ErrNotRegisteredFrameType {
		t.Errorf("I expected to get error \"%s\" but got \"%s\"", ErrNotRegisteredFrameType, err)
	}
}

func TestFrame_ProcessExistsHandler(t *testing.T) {
	handler := func(ctx *handlers.AppContext, f frame.Frame) (frame.Frame, error) {
		return frame.Frame{
			Type: "test",
			Data: f.Data,
		}, nil
	}

	router := NewFrame()
	router.AddHandler("ping", handler)

	ctx := &handlers.AppContext{}
	f := frame.Frame{
		Type: "ping",
		Data: "Yo",
	}

	answer, err := router.Process(ctx, f)
	if err != nil {
		t.Errorf("I got unexpected error: %s", err)
	}

	if answer.Type != "test" {
		t.Errorf("I expected to get frame with type \"test\" but got \"%s\"", answer.Type)
	}

	if answer.Data != f.Data {
		t.Errorf("I expected to get frame with data \"%s\" but got \"%s\"", f.Data, answer.Data)
	}
}
