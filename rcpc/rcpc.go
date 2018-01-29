package rcpc
//reflect chan procedure call

import (
	"errors"
	"reflect"
	"sync"
)

type Agent interface {
	Send(string, ...interface{}) error
	Call(string, ...interface{}) ([]interface{}, error)
}

const chan_cache_len = 256

type call_ret struct {
	val []interface{}
	err error
}

type message struct {
	cmd string
	arg []reflect.Value
	ret chan call_ret
}

type Server struct {
	self reflect.Value
	c    chan message
	wait sync.WaitGroup
}

func (r *Server) Close () {
	close(r.c)
	r.wait.Wait()
}

func do(f reflect.Value, v []reflect.Value) (ret []interface{}, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = errors.New(rec.(string))
		}
	}()
	ret = getInterface(f.Call(v))
	return
}

func getValue(in []interface{}) []reflect.Value {
    l := len(in)
    ret := make([]reflect.Value, l)
    for i:=0; i < l; i++ {
        ret[i] = reflect.ValueOf(in[i])
    }
    return ret
}

func getInterface(in []reflect.Value) []interface{} {
    l := len(in)
    ret := make([]interface{}, l)
    for i:=0; i < l; i++ {
        ret[i] = in[i].Interface()
    }
    return ret
}

func NewServer (self interface{}) *Server {
    r := new(Server)
	r.self = reflect.ValueOf(self)
	r.c = make(chan message, chan_cache_len)
	r.wait.Add(1)

	go func () {
        defer r.wait.Done()
        for val := range r.c {
            f := r.self.MethodByName(val.cmd)
            var ret []interface{}
            var err error
            if f.IsValid() {
                ret, err = do(f, val.arg)
            } else {
                err = errors.New("not find func:" + val.cmd)
            }

            if val.ret != nil {
                val.ret <- call_ret{ret, err}
                close(val.ret)
            }
        }
    }()
    return r
}

func (r *Server) Send(cmd string, arg ...interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	r.c <- message{cmd, getValue(arg), nil}
	return
}

func (r *Server) Call(cmd string, arg ...interface{}) (ret []interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	ret_c := make(chan call_ret)

	r.c <- message{cmd, getValue(arg), ret_c}

	ret_val := <-ret_c
	ret = ret_val.val
	err = ret_val.err
	return
}
