package bucket_test

import (
	"fmt"
	"hyperchain/hyperdb"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"hyperchain/tree/bucket"
	"math/big"
)

var (
	testStateImplName = "testStateImpl"
	logger = logging.MustGetLogger("bucket_test")
	configs map[string]interface{}
)
func init(){

	configs = viper.GetStringMap("ledger.state.dataStructure.configs")
}
// State structure for maintaining world state.
// This encapsulates a particular implementation for managing the state persistence
// This is not thread safe
type State struct {
	currentBlockNum *big.Int
	StateImpl       bucket.BucketTree
	key_valueMap    bucket.K_VMap
	updateStateImpl bool
}

// NewState constructs a new State. This Initializes encapsulated state implementation
func NewState() *State {
	logger.Infof("Initializing state implementation [%s]", testStateImplName)
	stateImpl := bucket.NewBucketTree(testStateImplName)
	err := stateImpl.Initialize(configs)
	if err != nil {
		panic(fmt.Errorf("Error during initialization of state implementation: %s", err))
	}
	return &State{big.NewInt(1),*stateImpl, make(map[string][]byte),false}
}

// TODO test
// set the Key_value map to the state
func (state *State) SetK_VMap(key_valueMap bucket.K_VMap,blockNum *big.Int){
	if(state.key_valueMap != nil){
		logger.Debugf("the state has key_valueMap,overwrite it")
	}
	state.key_valueMap = key_valueMap
	state.updateStateImpl = true
	state.currentBlockNum = blockNum
}

// TODO test
func (state *State) GetHash() ([]byte,error){
	logger.Debug("Enter - GetHash()")
	if state.updateStateImpl {
		logger.Debug("udpateing stateImpl with working-set")
		state.StateImpl.PrepareWorkingSet(state.key_valueMap,state.currentBlockNum)
		state.updateStateImpl = false
	}
	hash,err := state.StateImpl.ComputeCryptoHash()
	if err != nil {
		return nil,err
	}
	logger.Debug("Exit GetHash()")
	return hash,nil
}

// TODO test
// AddChangesForPersistence adds key-value pairs to writeBatch
func (state *State) AddChangesForPersistence(writeBatch hyperdb.Batch) {
	logger.Debug("state.addChangesForPersistence()...start")
	if state.updateStateImpl {
		state.StateImpl.PrepareWorkingSet(state.key_valueMap,state.currentBlockNum)
		state.updateStateImpl = false
	}
	state.StateImpl.AddChangesForPersistence(writeBatch)
	// TODO should add the metadata to the writeBatch?
	logger.Debug("state.addChangesForPersistence()...finished")
}

// TODO test
// CommitStateDelta commits the changes from state.ApplyStateDelta to the
// DB.
func (state *State) CommitStateDelta(writeBatch hyperdb.Batch) error {
	if state.updateStateImpl {
		state.StateImpl.PrepareWorkingSet(state.key_valueMap,state.currentBlockNum)
		state.updateStateImpl = false
	}
	state.StateImpl.AddChangesForPersistence(writeBatch)
	return writeBatch.Write()
}

func (state *State) Reset(changePersists bool){
	state.StateImpl.Reset()
}
