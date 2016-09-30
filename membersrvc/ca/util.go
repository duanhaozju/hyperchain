// author: Lizhong kuang
// date: 2016-09-29

package ca

import (
	"errors"
	"io"
	"log"
	mrand "math/rand"
	"time"

	pb "hyperchain/membersrvc/protos"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var (
	// Trace is a trace logger.
	Trace *log.Logger
	// Info is an info logger.
	Info *log.Logger
	// Warning is a warning logger.
	Warning *log.Logger
	// Error is an error logger.
	Error *log.Logger
	// Panic is a panic logger.
	Panic *log.Logger
)

// LogInit initializes the various loggers.
//
func LogInit(trace, info, warning, error, panic io.Writer) {
	Trace = log.New(trace, "TRACE: ", log.LstdFlags|log.Lshortfile)
	Info = log.New(info, "INFO: ", log.LstdFlags)
	Warning = log.New(warning, "WARNING: ", log.LstdFlags|log.Lshortfile)
	Error = log.New(error, "ERROR: ", log.LstdFlags|log.Lshortfile)
	Panic = log.New(panic, "PANIC: ", log.LstdFlags|log.Lshortfile)
}

var rnd = mrand.NewSource(time.Now().UnixNano())

func randomString(n int) string {
	b := make([]byte, n)

	for i, cache, remain := n-1, rnd.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rnd.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

//
// MemberRoleToString converts a member role representation from int32 to a string,
// according to the Role enum defined in ca.proto.
//
func MemberRoleToString(role pb.Role) (string, error) {
	roleMap := pb.Role_name

	roleStr := roleMap[int32(role)]
	if roleStr == "" {
		return "", errors.New("Undefined user role passed.")
	}

	return roleStr, nil
}
