package rcpc

import (
	"errors"
	"net/rpc"
	"reflect"
	"sync"
)

type message struct {
	Seq           uint64
	ServiceMethod string
	Error         string
	argv          reflect.Value
	replyv        reflect.Value
	mtype         *methodType
	next          *message
}

type ClientCodec struct {
	write_c chan *message
	read_c  chan *message

	*service
	temp    *message
	freemsg *message
	msgLock *sync.Mutex
	wg      sync.WaitGroup
}

var ErrNotFindMethod = errors.New("not find method")
var ErrArgvType = errors.New("argv type Error")
var ErrReplyvType = errors.New("replyv type Error")
var ErrClosedChannel = errors.New("channel is closed")
var ErrReplyCantSet = errors.New("reply value cant set")

const channel_len = 8192

func NewClient(rcvr interface{}) (*rpc.Client, error) {
	codec, err := NewClientCodec(rcvr)
	if err != nil {
		return nil, err
	}

	return rpc.NewClientWithCodec(codec), nil
}

func NewClientCodec(rcvr interface{}) (rpc.ClientCodec, error) {
	s := new(ClientCodec)
	s.write_c = make(chan *message, channel_len)
	s.read_c = make(chan *message, channel_len)
	s.msgLock = new(sync.Mutex)
	var err error
	s.service, err = newservice(rcvr)
	if err != nil {
		return nil, err
	}

	s.wg.Add(1)
	go s.run()
	return s, nil
}
func (s *ClientCodec) getmessage() *message {
	s.msgLock.Lock()
	msg := s.freemsg
	if msg == nil {
		msg = new(message)
	} else {
		s.freemsg = msg.next
		*msg = message{}
	}
	s.msgLock.Unlock()
	return msg
}

func (s *ClientCodec) freemessage(msg *message) {
	s.msgLock.Lock()
	msg.next = s.freemsg
	s.freemsg = msg
	s.msgLock.Unlock()
}

func (s *ClientCodec) run() {
	for m := range s.write_c {
		returnValues := m.mtype.method.Func.Call([]reflect.Value{s.rcvr, m.argv, m.replyv})

		errInter := returnValues[0].Interface()
		m.Error = ""
		if errInter != nil {
			m.Error = errInter.(error).Error()
		}
		s.read_c <- m
	}
	close(s.read_c)
	s.wg.Done()
}

// WriteRequest must be safe for concurrent use by multiple goroutines.
func (c *ClientCodec) WriteRequest(req *rpc.Request, argv interface{}) (err error) {
	mtype, ok := c.service.method[req.ServiceMethod]
	if !ok {
		err = ErrNotFindMethod
		return
	}

	if reflect.TypeOf(argv) != mtype.ArgType {
		err = ErrArgvType
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = ErrClosedChannel
		}
	}()

	replyv := reflect.New(mtype.ReplyType.Elem())

	switch mtype.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(mtype.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(mtype.ReplyType.Elem(), 0, 0))
	}
	msg := c.getmessage()
	msg.Seq = req.Seq
	msg.argv = reflect.ValueOf(argv)
	msg.ServiceMethod = req.ServiceMethod
	msg.mtype = mtype
	msg.replyv = replyv
	c.write_c <- msg
	return nil
}
func (c *ClientCodec) ReadResponseHeader(rep *rpc.Response) error {
	var ok bool
	c.temp, ok = <-c.read_c
	if !ok {
		return ErrClosedChannel
	}
	rep.Error = c.temp.Error
	rep.Seq = c.temp.Seq
	rep.ServiceMethod = c.temp.ServiceMethod
	return nil
}
func (c *ClientCodec) ReadResponseBody(reply interface{}) error {
	v := reflect.ValueOf(reply)
	if !v.IsValid() || v.IsNil() {
		return nil
		return ErrReplyCantSet
	}
	if v.Type() != c.temp.mtype.ReplyType {
		return nil
		return ErrReplyvType
	}

	if v.Elem().CanSet() {
		v.Elem().Set(c.temp.replyv.Elem())
	}
	c.freemessage(c.temp)
	return nil
}

func (c *ClientCodec) Close() error {
	close(c.write_c)
	c.wg.Wait()
	return nil
}
