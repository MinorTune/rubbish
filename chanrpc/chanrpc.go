package chanrpc

import (
	"errors"
	"log"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

const rpc_chan_len = 8192

var type_method *sync.Map
var call_pool *sync.Pool
var send_pool *sync.Pool

func init() {
	type_method = new(sync.Map)
	call_pool = new(sync.Pool)
	call_pool.New = func() interface{} {
		return &Call{Done: make(chan *Call, 1)}
	}
	send_pool = new(sync.Pool)
	send_pool.New = func() interface{} { return new(Call) }

}

type Call struct {
	Cmd    string
	Argv   interface{}
	Replyv interface{}
	Error  error
	Done   chan *Call
	refarg []reflect.Value
	method *methodType
}

func (c *Call) done() {
	if c.Done != nil {
		c.Done <- c
	} else {
		send_pool.Put(c)
	}
}

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

type Server struct {
	msg    chan *Call
	rcvr   reflect.Value
	method map[string]*methodType
	wg     *sync.WaitGroup
}

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

var ErrServerClosed = errors.New("cpc server is closed")
var ErrMethodNotRegister = errors.New("cmd not register")
var ErrMethodNotSend = errors.New("cmd need reply value")
var ErrMethodArgType = errors.New("arg type error")
var ErrMethodReplyType = errors.New("reply type error")
var ErrMethodRunTime = errors.New("cmd runtime error")

func NewServer(sev interface{}) *Server {
	rcvr := reflect.ValueOf(sev)
	rctype := rcvr.Type()
	mtypeinter, ok := type_method.Load(rctype)

	s := &Server{msg: make(chan *Call, rpc_chan_len), rcvr: rcvr, wg: new(sync.WaitGroup)}
	if ok {
		s.method = mtypeinter.(map[string]*methodType)
	} else {
		s.method = suitableMethods(rctype, false)
		type_method.Store(rctype, s.method)
	}

	s.wg.Add(1)
	go s.run()
	return s
}

func do(call *Call) {
	defer func() {
		if e := recover(); e != nil {
			call.Error = ErrMethodRunTime
		}
		call.done()
	}()

	returnValues := call.method.method.Func.Call(call.refarg)
	if len(returnValues) != 0 {
		if err := returnValues[0].Interface(); err != nil {
			call.Error = err.(error)
		}
	}
}

func (s *Server) run() {
	for v := range s.msg {
		do(v)
	}
	s.wg.Done()
}

func (s *Server) Send(cmd string, arg interface{}) (err error) {
	mtype, ok := s.method[cmd]
	if !ok {
		err = ErrMethodNotRegister
		return
	}

	argrv := reflect.ValueOf(arg)
	if argrv.Type() != mtype.ArgType {
		err = ErrMethodArgType
		return
	}

	if mtype.ReplyType != nil {
		err = ErrMethodNotSend
		return
	}

	call := send_pool.Get().(*Call)
	call.Cmd = cmd
	call.Argv = arg
	call.refarg = []reflect.Value{s.rcvr, argrv}
	call.method = mtype

	defer func() {
		if e := recover(); e != nil {
			err = ErrServerClosed
			send_pool.Put(call)
		}
	}()
	s.msg <- call
	return
}

func (s *Server) Call(cmd string, arg, reply interface{}) (err error) {
	call := <-s.Go(cmd, arg, reply).Done
	err = call.Error
	call_pool.Put(call)
	return
}

func (s *Server) Go(cmd string, arg, reply interface{}) (call *Call) {
	call = call_pool.Get().(*Call)
	defer func() {
		if e := recover(); e != nil {
			call.Error = ErrServerClosed
		}
		if call.Error != nil {
			call.done()
		}
	}()

	mtype, ok := s.method[cmd]
	if !ok {
		call.Error = ErrMethodNotRegister
		return
	}

	argrv := reflect.ValueOf(arg)
	if !argrv.IsValid() || argrv.Type() != mtype.ArgType || !argrv.Elem().IsValid() {
		call.Error = ErrMethodArgType
		return
	}

	reprv := reflect.ValueOf(reply)
	if !reprv.IsValid() || reprv.Type() != mtype.ReplyType || !reprv.Elem().CanSet() {
		call.Error = ErrMethodReplyType
		return
	}

	call.Cmd = cmd
	call.Argv = arg
	call.Replyv = reply
	call.refarg = []reflect.Value{s.rcvr, argrv, reprv}
	call.method = mtype

	s.msg <- call
	return call
}

func (s *Server) Close() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	close(s.msg)
	s.wg.Wait()
	return
}

//from goroot/src/net/rpc/
// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

// suitableMethods returns suitable Rpc methods of typ, it will report
// error using log if reportErr is true.
func suitableMethods(typ reflect.Type, reportErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		// Method needs three ins: receiver, *args, *reply.
		var replyType reflect.Type
		if mtype.NumIn() == 3 {
			// Second arg must be a pointer.
			replyType = mtype.In(2)
			if replyType.Kind() != reflect.Ptr {
				if reportErr {
					log.Printf("rpc.Register: reply type of method %q is not a pointer: %q\n", mname, replyType)
				}
				continue
			}
			// Reply type must be exported.
			if !isExportedOrBuiltinType(replyType) {
				if reportErr {
					log.Printf("rpc.Register: reply type of method %q is not exported: %q\n", mname, replyType)
				}
				continue
			}
			// Method needs one out.
			if mtype.NumOut() != 1 {
				if reportErr {
					log.Printf("rpc.Register: method %q has %d output parameters; needs exactly one\n", mname, mtype.NumOut())
				}
				continue
			}
			// The return type of the method must be error.
			if returnType := mtype.Out(0); returnType != typeOfError {
				if reportErr {
					log.Printf("rpc.Register: return type of method %q is %q, must be error\n", mname, returnType)
				}
				continue
			}
		} else if mtype.NumIn() != 2 {
			log.Printf("rpc.Register: method %q has %d input parameters; needs exactly three\n", mname, mtype.NumIn())
			continue
		}
		// First arg need not be a pointer.
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			if reportErr {
				log.Printf("rpc.Register: argument type of method %q is not exported: %q\n", mname, argType)
			}
			continue
		}
		methods[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}
