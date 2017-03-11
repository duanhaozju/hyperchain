package main

import (
	"fmt"
	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
	"hyperchain/api/jsonrpc/core"
	"hyperchain/crypto/hmEncryption"
	"hyperchain/tree/bucket"
	"math/big"
	"time"
	"hyperchain/common"
	"os/exec"
)

const HyperchainVersion ="Hyperchain Version:\nRelease1.2\n"

func GetOperationSystem()(string,error){
	f, err := exec.Command("lsb_release", "-a").Output()
	if err != nil {
		return "",err
	}
	return string(f),nil
}

func GetHyperchainVersion()string{
	return HyperchainVersion
}
type configs interface {
	initConfig(NodeID int, GRPCPort int, HTTPPort int)
	getNodeID() int
	getGRPCPort() int
	getHTTPPort() int
	getKeystoreDir() string
	getLogDumpFileFlag() bool
	getLogDumpFileDir() string
	getLogLevel() string
	getDatabaseDir() string
	getPeerConfigPath() string
	getGenesisConfigPath() string
	getMemberSRVCConfigPath() string
	getPBFTConfigPath() string
	getSyncReplicaInterval() (time.Duration, error)
	getSyncReplicaEnable() bool
	getLicense() string
	getRateLimitConfig() jsonrpc.RateLimitConfig
	getStateType() string
	getBlockVersion() string
	getTransactionVersion() string
	getPaillerPublickey() hmEncryption.PaillierPublickey
	getBucketTreeConf() bucket.Conf
	getDbConfig() string
}

type configsImpl struct {
	nodeID               int
	gRPCPort             int
	httpPort             int
	restPort             int
	keystoreDir          string
	keyNodeDir           string
	logDumpFileFlag      bool
	logDumpFileDir       string
	logLevel             string
	dbConfig                string
	databaseDir          string
	peerConfigPath       string
	genesisConfigPath    string
	memberSRVCConfigPath string
	pbftConfigPath       string
	// sync replica info
	syncReplicaInfoInterval string
	syncReplica             bool
	// license
	license string
	// rate limit related
	rateLimitEnable  bool
	txRatePeak       int64
	txFillRate       string
	contractRatePeak int64
	contractFillRate string
	// data structure version related
	blockVersion       string
	transactionVersion string
	// state type
	stateType string
	// bucket tree related
	stateSize             int // state db bucket tree size
	stateLevelGroup       int // state db bucket tree level group
	storageSize           int // storage bucket tree size
	storageLevelGroup     int // storage bucket tree level group
	paillpublickeyN       string
	paillpublickeynsquare string
	paillpublickeyG       string
}

//return a config instances
func newconfigsImpl(conf *common.Config) *configsImpl {
//func newconfigsImpl(globalConfigPath string, NodeID int, GRPCPort int, HTTPPort int, RESTPort int) *configsImpl {
	var cimpl configsImpl
	config := viper.New()
	//viper.SetEnvPrefix("GLOBAL_ENV")
	peerConfigPath := conf.GetString(common.C_GLOBAL_CONFIG_PATH)
	//fmt.Println(peerConfigPath)
	config.SetConfigFile(peerConfigPath)
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error envPre %s reading %s", "GLOBAL", err))
	}
	/*
		system config
	*/
	cimpl.nodeID = conf.GetInt(common.C_NODE_ID)
	cimpl.gRPCPort = conf.GetInt(common.C_GRPC_PORT)
	cimpl.httpPort = conf.GetInt(common.C_HTTP_PORT)
	cimpl.restPort = conf.GetInt(common.C_REST_PORT)
	cimpl.keystoreDir = config.GetString("global.account.keystoredir")
	cimpl.keyNodeDir = config.GetString("global.account.keynodesdir")
	cimpl.logDumpFileFlag = config.GetBool("global.logs.dumpfile")
	cimpl.logDumpFileDir = config.GetString("global.logs.logsdir")
	cimpl.logLevel = config.GetString("global.logs.loglevel")
	cimpl.databaseDir = config.GetString("global.database.dir")
	cimpl.peerConfigPath = config.GetString("global.configs.peers")
	cimpl.genesisConfigPath = config.GetString("global.configs.genesis")
	cimpl.memberSRVCConfigPath = config.GetString("global.configs.membersrvc")
	cimpl.pbftConfigPath = config.GetString("global.configs.pbft")
	/*
		db Config
	*/
	cimpl.dbConfig = config.GetString("global.dbConfig")
	/*
		statement synchronization
	*/
	cimpl.syncReplicaInfoInterval = config.GetString("global.configs.replicainfo.interval")
	cimpl.syncReplica = config.GetBool("global.configs.replicainfo.enable")
	/*
		license
	*/
	cimpl.license = config.GetString("global.configs.license")
	/*
		rate limit
	*/
	cimpl.rateLimitEnable = config.GetBool("global.configs.ratelimit.enable")
	cimpl.txRatePeak = config.GetInt64("global.configs.ratelimit.txRatePeak")
	cimpl.txFillRate = config.GetString("global.configs.ratelimit.txFillRate")
	cimpl.contractRatePeak = config.GetInt64("global.configs.ratelimit.contractRatePeak")
	cimpl.contractFillRate = config.GetString("global.configs.ratelimit.contractFillRate")
	cimpl.stateType = config.GetString("global.structure.state")

	/*
		Version
	*/
	cimpl.blockVersion = config.GetString("global.version.blockversion")
	cimpl.transactionVersion = config.GetString("global.version.transactionversion")

	cimpl.paillpublickeyN = config.GetString("global.configs.hmpublickey.N")
	cimpl.paillpublickeynsquare = config.GetString("global.configs.hmpublickey.Nsquare")
	cimpl.paillpublickeyG = config.GetString("global.configs.hmpublickey.G")

	/*
		Bucket tree
	*/
	cimpl.stateSize = config.GetInt("global.configs.buckettree.state.size")
	cimpl.stateLevelGroup = config.GetInt("global.configs.buckettree.state.levelGroup")
	cimpl.storageSize = config.GetInt("global.configs.buckettree.storage.size")
	cimpl.storageLevelGroup = config.GetInt("global.configs.buckettree.storage.levelGroup")
	return &cimpl
}

func (cIml *configsImpl) getNodeID() int            { return cIml.nodeID }
func (cIml *configsImpl) getGRPCPort() int          { return cIml.gRPCPort }
func (cIml *configsImpl) getHTTPPort() int          { return cIml.httpPort }
func (cIml *configsImpl) getRESTPort() int          { return cIml.restPort }
func (cIml *configsImpl) getDbConfig() string       { return cIml.dbConfig }
func (cIml *configsImpl) getKeystoreDir() string    { return cIml.keystoreDir }
func (cIml *configsImpl) getKeyNodeDir() string     { return cIml.keyNodeDir }
func (cIml *configsImpl) getLogDumpFileFlag() bool  { return cIml.logDumpFileFlag }
func (cIml *configsImpl) getLogDumpFileDir() string { return cIml.logDumpFileDir }
func (cIml *configsImpl) getLogLevel() logging.Level {
	switch cIml.logLevel {
	case "CRITICAL":
		return logging.CRITICAL
	case "ERROR":
		return logging.ERROR
	case "WARNING":
		return logging.WARNING
	case "NOTICE":
		return logging.NOTICE
	case "INFO":
		return logging.INFO
	case "DEBUG":
		return logging.DEBUG
	default:
		return logging.NOTICE
	}
}
func (cIml *configsImpl) getDatabaseDir() string       { return cIml.databaseDir }
func (cIml *configsImpl) getPeerConfigPath() string    { return cIml.peerConfigPath }
func (cIml *configsImpl) getGenesisConfigPath() string { return cIml.genesisConfigPath }
func (cIml *configsImpl) getMemberSRVCConfigPath() string {
	return cIml.memberSRVCConfigPath
}
func (cIml *configsImpl) getPBFTConfigPath() string { return cIml.pbftConfigPath }
func (cIml *configsImpl) getSyncReplicaInterval() (time.Duration, error) {
	return time.ParseDuration(cIml.syncReplicaInfoInterval)
}
func (cIml *configsImpl) getSyncReplicaEnable() bool { return cIml.syncReplica }
func (cIml *configsImpl) getLicense() string         { return cIml.license }
func (cIml *configsImpl) getRateLimitConfig() jsonrpc.RateLimitConfig {
	txFillRate, _ := time.ParseDuration(cIml.txFillRate)
	contractFillRate, _ := time.ParseDuration(cIml.contractFillRate)
	return jsonrpc.RateLimitConfig{
		Enable:           cIml.rateLimitEnable,
		TxRatePeak:       cIml.txRatePeak,
		TxFillRate:       txFillRate,
		ContractRatePeak: cIml.contractRatePeak,
		ContractFillRate: contractFillRate,
	}
}
func (cIml *configsImpl) getStateType() string { return cIml.stateType }
func (cIml *configsImpl) getBlockVersion() string {
	return cIml.blockVersion
}
func (cIml *configsImpl) getTransactionVersion() string {
	return cIml.transactionVersion
}

func (cIml *configsImpl) getBucketTreeConf() bucket.Conf {
	return bucket.Conf{
		StateSize:         cIml.stateSize,
		StateLevelGroup:   cIml.stateLevelGroup,
		StorageSize:       cIml.storageSize,
		StorageLevelGroup: cIml.storageLevelGroup,
	}
}

func (cIml *configsImpl) getPaillerPublickey() *hmEncryption.PaillierPublickey {
	bigN := new(big.Int)
	bigNsquare := new(big.Int)
	bigG := new(big.Int)
	n, _ := bigN.SetString(cIml.paillpublickeyN, 10)
	nsquare, _ := bigNsquare.SetString(cIml.paillpublickeynsquare, 10)
	g, _ := bigG.SetString(cIml.paillpublickeyG, 10)

	return &hmEncryption.PaillierPublickey{
		N:       n,
		Nsquare: nsquare,
		G:       g,
	}
}
