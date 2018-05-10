package chanrpc

import (
	"errors"
	"reflect"
	"sync"
)

var rpc_server_closed = errors.New("rpc server is closed")
var command_not_register = errors.New("rpc server command not register")
var rpc_runtime_panic = errors.New("rpc server runtime panic")

var call_pool *sync.Pool
var methods_map *sync.Map

func init() {
	call_pool = new(sync.Pool)
	call_pool.New = func() interface{} {
		return &Call{}
	}

	methods_map = new(sync.Map)
}

type Call struct {
	method *reflect.Method
	args   []reflect.Value
	err    error
	c      chan struct{}
}

func (call *Call) Done() (err error) {
	<-call.c
	err = call.err
	call_pool.Put(call)
	return
}

func (call *Call) done() {
	if call.c != nil {
		close(call.c)
	} else {
		call_pool.Put(call)
	}
}

func (s *Server) putCall(call *Call) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = rpc_server_closed
			call.done()
		}
	}()
	s.call_store <- call
	return
}

type Server struct {
	method  map[string]*reflect.Method
	self_rv reflect.Value

	call_store chan *Call
	wt         *sync.WaitGroup
}

func NewRpc(r interface{}) *Server {
	s := new(Server)
	s.self_rv = reflect.ValueOf(r)
	s.call_store = make(chan *Call, 256)

	self_typ := s.self_rv.Type()
	methods, ok := methods_map.Load(self_typ)
	if ok {
		s.method = methods.(map[string]*reflect.Method)
	} else {
		s.method = suitableMethods(self_typ)
		methods_map.Store(self_typ, s.method)
	}

	s.wt = new(sync.WaitGroup)
	s.wt.Add(1)
	go func() {
		for c := range s.call_store {
			docall(c)
		}

		s.wt.Done()
	}()
	return s
}

func docall(call *Call) {
	defer func() {
		if e := recover(); e != nil {
			if str, ok := e.(string); ok {
				call.err = errors.New(str)
			} else if err, ok := e.(error); ok {
				call.err = err
			} else {
				call.err = rpc_runtime_panic
			}
		}
		call.done()
	}()
	call.method.Func.Call(call.args)
}

func (s *Server) Send(cmd string, args ...interface{}) (err error) {
	method, ok := s.method[cmd]
	if !ok {
		err = command_not_register
		return
	}

	call := call_pool.Get().(*Call)
	call.c = nil

	call.method = method

	args_len := len(args)
	call.args = make([]reflect.Value, args_len+1)
	call.args[0] = s.self_rv
	for i := 0; i < args_len; i++ {
		call.args[i+1] = reflect.ValueOf(args[i])
	}

	err = s.putCall(call)
	return
}
func (s *Server) Call(cmd string, args ...interface{}) error {
	return s.Go(cmd, args...).Done()
}
func (s *Server) Go(cmd string, args ...interface{}) (call *Call) {
	call = call_pool.Get().(*Call)
	call.c = make(chan struct{})

	var ok bool
	call.method, ok = s.method[cmd]
	if !ok {
		call.err = command_not_register
		call.done()
		return
	}

	args_len := len(args)
	call.args = make([]reflect.Value, args_len+1)
	call.args[0] = s.self_rv
	for i := 0; i < args_len; i++ {
		call.args[i+1] = reflect.ValueOf(args[i])
	}

	err := s.putCall(call)
	if err != nil {
		call.err = err
		call.done()
	}
	return
}
func suitableMethods(typ reflect.Type) map[string]*reflect.Method {
	methods := make(map[string]*reflect.Method)

	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}

		if mtype.NumOut() == 0 {
			methods[mname] = &method
		}
	}
	return methods
}
