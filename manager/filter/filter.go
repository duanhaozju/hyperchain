package filter

import (
	"time"
	"hyperchain/common"
	"hyperchain/core/vm"
)

var (
	deadline = 5 * time.Minute // consider a filter inactive if it has not been polled for within deadline
)

// filter is a helper struct that holds meta information over the filter type
// and associated subscription in the event system.
type Filter struct {
	typ      Type
	deadline *time.Timer // filter is inactiv when deadline triggers
	hashes   []common.Hash
	crit     FilterCriteria
	logs     []*vm.Log
	s        *Subscription // associated subscription in event system
}

func NewFilter(typ Type, sub *Subscription, crit FilterCriteria) *Filter {
	return &Filter{
		typ:      typ,
		deadline: time.NewTimer(deadline),
		s:        sub,
		crit:     crit,
	}
}

/*
	Getter
 */
func (flt *Filter) GetDeadLine() *time.Timer {
	return flt.deadline
}

func (flt *Filter) GetHashes() []common.Hash {
	return flt.hashes
}

func (flt *Filter) GetLogs() []*vm.Log {
	return flt.logs
}

func (flt *Filter) GetSubsctiption() *Subscription {
	return flt.s
}

func (flt *Filter) GetType() Type {
	return flt.typ
}

func (flt *Filter) GetCriteria() FilterCriteria {
	return flt.crit
}

/*
	Setter
 */

func (flt *Filter) AddHash(hash common.Hash) {
	flt.hashes = append(flt.hashes, hash)
}

func (flt *Filter) ClearHash() {
	flt.hashes = nil
}

func (flt *Filter) AddLog(log *vm.Log) {
	flt.logs = append(flt.logs, log)
}

func (flt *Filter) Clearlog() {
	flt.logs = nil
}

func (flt *Filter) ResetDeadline() {
	flt.deadline.Reset(deadline)
}
