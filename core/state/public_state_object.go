package state

import (
	"time"
	"hyperchain/common"
)
var PublicStateObjectMaps PublicStateObjectMap
type PublicStateObject struct{
	State_object	*StateObject
	Frequence	int
	Timestamp	time.Time
}

type PublicStateObjectMap struct{
	PublicStateObjects	map[common.Address]PublicStateObject
}

func Init(){
	PublicStateObjectMaps = PublicStateObjectMap{}
}
