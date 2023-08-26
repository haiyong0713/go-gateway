package request_test

import (
	"net/http"
	"testing"

	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"

	"github.com/stretchr/testify/assert"
)

func TestHandlerList(t *testing.T) {
	s := ""
	r := &request.Request{}
	l := request.HandlerList{}
	l.PushBack(func(r *request.Request) {
		s += "a"
		r.Data = s
	})
	l.Run(r)
	assert.Equal(t, "a", r.Data.(string))
}

func TestMultipleHandlers(t *testing.T) {
	r := &request.Request{}
	l := request.HandlerList{}
	l.PushBack(func(r *request.Request) {
		r.Data = nil
	})
	l.PushFront(func(r *request.Request) {
		r.Data = rootsdk.Bool(true)
	})
	l.Run(r)
	assert.Empty(t, r.Data)
}

func TestNamedHandlers(t *testing.T) {
	l := request.HandlerList{}
	named1 := request.NamedHandler{Name: "Name1", Fn: func(r *request.Request) {}}
	named2 := request.NamedHandler{Name: "Name2", Fn: func(r *request.Request) {}}
	l.PushBackNamed(named1)
	l.PushBackNamed(named1)
	l.PushBackNamed(named2)
	l.PushBack(func(r *request.Request) {})
	assert.Equal(t, 4, l.Len())
	l.Remove(named2)
	assert.Equal(t, 3, l.Len())
	l.Remove(named1)
	assert.Equal(t, 1, l.Len())
}

func TestSwapHandlers(t *testing.T) {
	firstHandlerCalled := 0
	swappedOutHandlerCalled := 0
	swappedInHandlerCalled := 0
	l := request.HandlerList{}
	named1 := request.NamedHandler{Name: "Name", Fn: func(r *request.Request) {
		firstHandlerCalled++
	}}
	named2 := request.NamedHandler{Name: "SwapOutName", Fn: func(r *request.Request) {
		swappedOutHandlerCalled++
	}}
	l.PushBackNamed(named1)
	l.PushBackNamed(named2)
	l.PushBackNamed(named1)
	l.SwapNamed(request.NamedHandler{Name: "SwapOutName", Fn: func(r *request.Request) {
		swappedInHandlerCalled++
	}})
	l.Run(&request.Request{})
	assert.Equal(t, 2, firstHandlerCalled)
	assert.Equal(t, 1, swappedInHandlerCalled)
	assert.Equal(t, 0, swappedOutHandlerCalled)
}

func TestSetBackOrFrontNamed(t *testing.T) {
	firstHandlerCalled := 0
	secondHandlerCalled := 0
	swappedOutHandlerCalled := 0
	swappedInHandlerCalled := 0
	otherHandlerCalled := 0
	l := request.HandlerList{}
	named1 := request.NamedHandler{Name: "Name", Fn: func(r *request.Request) {
		firstHandlerCalled++
	}}
	named2 := request.NamedHandler{Name: "SwapOutName", Fn: func(r *request.Request) {
		swappedOutHandlerCalled++
	}}
	l.PushBackNamed(named1)
	l.PushBackNamed(named2)
	l.SetBackNamed(request.NamedHandler{Name: "SwapOutName", Fn: func(r *request.Request) {
		swappedInHandlerCalled++
	}})
	l.SetBackNamed(request.NamedHandler{Name: "OtherName", Fn: func(r *request.Request) {
		otherHandlerCalled++
	}})
	l.SetFrontNamed(request.NamedHandler{Name: "Name", Fn: func(r *request.Request) {
		secondHandlerCalled++
	}})
	l.SetFrontNamed(request.NamedHandler{Name: "OtherName1", Fn: func(r *request.Request) {
		otherHandlerCalled++
	}})
	l.Run(&request.Request{})
	assert.Equal(t, 0, firstHandlerCalled)
	assert.Equal(t, 1, swappedInHandlerCalled)
	assert.Equal(t, 2, otherHandlerCalled)
	assert.Equal(t, 0, swappedOutHandlerCalled)
	assert.Equal(t, 1, secondHandlerCalled)
}

func TestPushBackOrFront(t *testing.T) {
	l := request.HandlerList{}
	b := make([]byte, 0)
	named1 := request.NamedHandler{Name: "Name1", Fn: func(r *request.Request) {
		b = append(b, '1')
	}}
	named2 := request.NamedHandler{Name: "Name2", Fn: func(r *request.Request) {
		b = append(b, '2')
	}}
	l.PushBackNamed(named1)
	l.PushBackNamed(named2)
	l.PushFrontNamed(named2)
	l.Run(&request.Request{})
	assert.Equal(t, []byte{'2', '1', '2'}, b)
}

func TestStopHandlers(t *testing.T) {
	l := request.HandlerList{}
	stopAt := 1
	l.AfterEachFn = func(item request.HandlerListRunItem) bool {
		return item.Index != stopAt
	}
	called := 0
	l.PushBackNamed(request.NamedHandler{Name: "name1", Fn: func(r *request.Request) {
		called++
	}})
	l.PushBackNamed(request.NamedHandler{Name: "name2", Fn: func(r *request.Request) {
		called++
	}})
	l.PushBackNamed(request.NamedHandler{Name: "name3", Fn: func(r *request.Request) {
		t.Fatalf("third handler should not be called")
	}})
	l.Run(&request.Request{})
	assert.Equal(t, 2, called)
}

func BenchmarkHandlersCopy(b *testing.B) {
	handlers := request.Handlers{}

	handlers.Validate.PushBack(func(r *request.Request) {})
	handlers.Validate.PushBack(func(r *request.Request) {})
	handlers.Build.PushBack(func(r *request.Request) {})
	handlers.Build.PushBack(func(r *request.Request) {})
	handlers.Send.PushBack(func(r *request.Request) {})
	handlers.Send.PushBack(func(r *request.Request) {})
	handlers.Unmarshal.PushBack(func(r *request.Request) {})
	handlers.Unmarshal.PushBack(func(r *request.Request) {})

	for i := 0; i < b.N; i++ {
		h := handlers.Copy()
		assert.Equal(b, handlers.Validate.Len(), h.Validate.Len())
	}
}

func BenchmarkHandlersPushBack(b *testing.B) {
	handlers := request.Handlers{}
	for i := 0; i < b.N; i++ {
		h := handlers.Copy()
		h.Validate.PushBack(func(r *request.Request) {})
		h.Validate.PushBack(func(r *request.Request) {})
		h.Validate.PushBack(func(r *request.Request) {})
		h.Validate.PushBack(func(r *request.Request) {})
	}
}

func BenchmarkHandlersClear(b *testing.B) {
	handlers := request.Handlers{}

	for i := 0; i < b.N; i++ {
		h := handlers.Copy()
		h.Validate.PushFront(func(r *request.Request) {})
		h.Validate.PushFront(func(r *request.Request) {})
		h.Validate.PushFront(func(r *request.Request) {})
		h.Validate.PushFront(func(r *request.Request) {})
		h.Clear()
	}
}

func TestHandlersClear(t *testing.T) {
	h := request.Handlers{}
	h.Complete.PushFront(func(r *request.Request) {})
	assert.Equal(t, false, h.IsEmpty())
	h.Clear()
	assert.Equal(t, true, h.IsEmpty())
}

func TestWithSetRequestHeaders(t *testing.T) {
	h := make(map[string]string, 2)
	h["k1"] = "v1"
	h["k2"] = "v2"
	r := &request.Request{
		HTTPRequest: &http.Request{
			Header: http.Header{},
		},
	}
	opt := request.WithSetRequestHeaders(h)
	r.ApplyOptions(opt)
	assert.Equal(t, r.HTTPRequest.Header["k1"][0], "v1")
	assert.Equal(t, r.HTTPRequest.Header["k2"][0], "v2")
}
