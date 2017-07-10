package hts

import (
	"hyperchain/p2p/threadsafelinkedlist"
	"time"
)

type SessionKey struct {
	//if time.Now() > expire, this key should be updated
	expire    int64
	sharedKey []byte
	//this is thread safe flag
	valid *threadsafelinkedlist.SpinLock
}

func NewSessionKey(sharedKey []byte) *SessionKey {
	sk := &SessionKey{
		expire:    time.Now().Unix() + 3600,
		sharedKey: sharedKey,
		valid:     new(threadsafelinkedlist.SpinLock),
	}
	go sk.selfCheck()
	return sk
}

func (sk *SessionKey) GetKey() []byte {
	if sk.IsValid() {
		return sk.sharedKey
	}
	return nil
}

func (sk *SessionKey) Update(updateKey []byte) {
	if sk.IsValid() {
		sk.SetOff()
	}
	sk.sharedKey = updateKey
	//session key expire time is (now unix) + 3600s = 1hour
	sk.expire += time.Now().Unix() + 3600
	sk.SetOn()
}

func (sk *SessionKey) selfCheck() {
	for t := range time.Tick(10 * time.Second) {
		if t.Unix() > sk.expire {
			sk.SetOff()
		}
	}
}

func (sk *SessionKey) IsValid() bool {
	//if is locked, this key is not valid
	return !sk.valid.IsLocked()
}

func (sk *SessionKey) SetOn() {
	sk.valid.UnLock()
}

func (sk *SessionKey) SetOff() {
	sk.valid.TryLock()
}
