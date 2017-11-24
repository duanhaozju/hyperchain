package kvlog

import "errors"

var (
	ErrLidTooLarge = errors.New("lid is too large")

	ErrSeqNoTooLarge = errors.New("seqNo is too large")

	ErrNoLid = errors.New("cannot find lid in database")

	ErrNoSeqNo = errors.New("cannot find seqNo in database")

	ErrAppendFail = errors.New("cannot append to opLog")

	ErrMismatch = errors.New("unmatch seqNo in database")

	ErrMarshal = errors.New("cannot marshal")

	ErrUnmarshal = errors.New("cannot unmarshal")

)
