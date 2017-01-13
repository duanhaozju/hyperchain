package bucket_test

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"hyperchain/hyperdb"
	"hyperchain/tree/bucket"
	"math/big"
)

var (
	testStateImplName = "testStateImpl"
	logger            = logging.MustGetLogger("bucket_test")
	configs           map[string]interface{}
)

func init() {
	configs = viper.GetStringMap("ledger.state.dataStructure.configs")
}

// State structure for maintaining world state.
// This encapsulates a particular implementation for managing the state persistence
// This is not thread safe
type State struct {
	currentBlockNum *big.Int
	Bucket_tree     bucket.BucketTree
	key_valueMap    bucket.K_VMap
	updateStateImpl bool
}

// NewState constructs a new State. This Initializes encapsulated state implementation
func NewState(treePrefix string) *State {
	logger.Infof("Initializing state implementation [%s]", testStateImplName)
	stateImpl := bucket.NewBucketTree(treePrefix)
	err := stateImpl.Initialize(configs)
	if err != nil {
		panic(fmt.Errorf("Error during initialization of state implementation: %s", err))
	}
	return &State{big.NewInt(0), *stateImpl, make(map[string][]byte), false}
}

// TODO test
// set the Key_value map to the state
func (state *State) SetK_VMap(key_valueMap bucket.K_VMap, blockNum *big.Int) {
	if state.key_valueMap != nil {
		logger.Debugf("the state has key_valueMap,overwrite it")
	}
	state.key_valueMap = key_valueMap
	state.updateStateImpl = true
	state.currentBlockNum = blockNum
}

// TODO test
func (state *State) GetHash() ([]byte, error) {
	logger.Debug("Enter - GetHash()")
	if state.updateStateImpl {
		logger.Debug("udpateing stateImpl with working-set")
		state.Bucket_tree.PrepareWorkingSet(state.key_valueMap, state.currentBlockNum)
		state.updateStateImpl = false
	}
	hash, err := state.Bucket_tree.ComputeCryptoHash()
	if err != nil {
		return nil, err
	}
	logger.Debug("Exit GetHash()")
	return hash, nil
}

// TODO test
// AddChangesForPersistence adds key-value pairs to writeBatch
func (state *State) AddChangesForPersistence(writeBatch hyperdb.Batch, blockNum *big.Int) {
	logger.Debug("state.addChangesForPersistence()...start")
	if state.updateStateImpl {
		state.Bucket_tree.PrepareWorkingSet(state.key_valueMap, state.currentBlockNum)
		state.updateStateImpl = false
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch, blockNum)
	// TODO should add the metadata to the writeBatch?
	logger.Debug("state.addChangesForPersistence()...finished")
}

// TODO test
// CommitStateDelta commits the changes from state.ApplyStateDelta to the
// DB.
func (state *State) CommitStateDelta(writeBatch hyperdb.Batch, blockNum *big.Int) error {
	if state.updateStateImpl {
		state.Bucket_tree.PrepareWorkingSet(state.key_valueMap, state.currentBlockNum)
		state.updateStateImpl = false
	}
	state.Bucket_tree.AddChangesForPersistence(writeBatch, blockNum)
	return writeBatch.Write()
}

func (state *State) RevertToTargetBlock(currentBlockNum, toBlockNum *big.Int) error {
	logger.Debug("start to revert to target block")
	defer logger.Debug("end to revert to target block")
	return state.Bucket_tree.RevertToTargetBlock(currentBlockNum, toBlockNum)
}

func (state *State) Reset(changePersists bool) {
	state.Bucket_tree.Reset()
}
