package hts

import (
	"hyperchain/p2p/threadsafe"
	"time"
)

type SessionKey struct {
	expire    int64				// if time.Now() > expire, this key should be updated
	sharedKey []byte			// the shared session key
	valid *threadsafe.SpinLock
}

// NewSessionKey creates and returns a new SessionKey instance.
func NewSessionKey(sharedKey []byte) *SessionKey {
	sk := &SessionKey{
		expire:    time.Now().Unix() + 3600,
		sharedKey: sharedKey,
		valid:     new(threadsafe.SpinLock),
	}
	//TODO next version feature.
	//go sk.selfCheck()
	return sk
}

// GetKey returns the shared session key if it is valid,
// otherwise, returns nil.
func (sk *SessionKey) GetKey() []byte {
	if sk.IsValid() {
		return sk.sharedKey
	}
	return nil
}

// Update will update the shared session key and reset expiration time.
func (sk *SessionKey) Update(updateKey []byte) {
	if sk.IsValid() {
		sk.SetOff()
	}
	sk.sharedKey = updateKey

	// session key expire time is (now unix) + 3600s = 1hour
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

// IsValid returns true if unlocked.
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
