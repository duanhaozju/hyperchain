package filter

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/manager/event"
	"reflect"
	"time"
)

var (
	deadline = 5 * time.Minute // consider a filter inactive if it has not been polled for within deadline
)

// filter is a helper struct that holds meta information over the filter type
// and associated subscription in the event system.
type Filter struct {
	typ      Type
	deadline *time.Timer // filter is inactiv when deadline triggers
	data     []interface{}
	crit     FilterCriteria
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

func (flt *Filter) GetData() []interface{} {
	return flt.data
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

func (flt *Filter) AddData(d interface{}) {
	flt.data = append(flt.data, d)
}

func (flt *Filter) ClearData() {
	flt.data = nil
}

func (flt *Filter) ResetDeadline() {
	flt.deadline.Reset(deadline)
}

// filterLogs creates a slice of logs matching the given criteria.
func filterLogs(logs []*types.Log, logCrit *FilterCriteria) []*types.Log {
	var ret []*types.Log
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

func filterException(ev event.FilterSystemStatusEvent, crit *FilterCriteria) bool {
	include := func(include []interface{}, exclude []interface{}, val interface{}) bool {
		if len(include) == 0 {
			if len(exclude) == 0 {
				// if no include, exclude collection been specfied, just regard the crit as a wildcard
				return true
			} else {
				for _, e := range exclude {
					if reflect.DeepEqual(e, val) {
						return false
					}
				}
				return true
			}
		} else {
			// the effect of include is larger than exclude
			for _, e := range include {
				if reflect.DeepEqual(e, val) {
					return true
				}
			}
			return false
		}
	}
	convertStringArray := func(in []string) []interface{} {
		ret := make([]interface{}, len(in))
		for idx, val := range in {
			ret[idx] = val
		}
		return ret
	}
	convertIntArray := func(in []int) []interface{} {
		ret := make([]interface{}, len(in))
		for idx, val := range in {
			ret[idx] = val
		}
		return ret
	}
	if include(convertStringArray(crit.Modules), convertStringArray(crit.ModulesExclude), ev.Module) &&
		include(convertStringArray(crit.SubType), convertStringArray(crit.SubTypeExclude), ev.Subtype) &&
		include(convertIntArray(crit.Code), convertIntArray(crit.CodeExclude), ev.ErrorCode) {
		return true
	}
	return false
}

func includes(addresses []common.Address, a common.Address) bool {
	for _, addr := range addresses {
		if addr == a {
			return true
		}
	}
	return false
}
