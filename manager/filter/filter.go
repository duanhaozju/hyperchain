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

func (flt *Filter) GetVerbose() bool {
	return flt.s.f.verbose
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

func (flt *Filter) AddLog(log []*vm.Log) {
	flt.logs = append(flt.logs, log...)
}

func (flt *Filter) Clearlog() {
	flt.logs = nil
}

func (flt *Filter) ResetDeadline() {
	flt.deadline.Reset(deadline)
}

// filterLogs creates a slice of logs matching the given criteria.
func filterLogs(logs []*vm.Log, logCrit *FilterCriteria) []*vm.Log {
	var ret []*vm.Log
	Logs:
	for _, log := range logs {
		if logCrit.FromBlock != nil && logCrit.FromBlock.Int64() >= 0 && logCrit.FromBlock.Uint64() > log.BlockNumber {
			continue
		}
		if logCrit.ToBlock != nil && logCrit.ToBlock.Int64() >= 0 && logCrit.ToBlock.Uint64() < log.BlockNumber {
			continue
		}

		if len(logCrit.Addresses) > 0 && !includes(logCrit.Addresses, log.Address) {
			continue
		}

		// If the to filtered topics is greater than the amount of topics in logs, skip.
		if len(logCrit.Topics) > len(log.Topics) {
			continue Logs
		}

		for i, topics := range logCrit.Topics {
			var match bool
			for _, topic := range topics {
				// common.Hash{} is a match all (wildcard)
				if (topic == common.Hash{}) || log.Topics[i] == topic {
					match = true
					break
				}
			}

			if !match {
				continue Logs
			}
		}
		ret = append(ret, log)
	}

	return ret
}

func includes(addresses []common.Address, a common.Address) bool {
	for _, addr := range addresses {
		if addr == a {
			return true
		}
	}

	return false
}
