package main

import (
	"fmt"

	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
	"time"
	"hyperchain/api/jsonrpc/core"
)

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
	getBlockVersion() string
	getTransactionVersion() string
}

type configsImpl struct {
	nodeID                  int
	gRPCPort                int
	httpPort                int
	restPort		int
	keystoreDir             string
	keyNodeDir              string
	logDumpFileFlag         bool
	logDumpFileDir          string
	logLevel                string
	databaseDir             string
	peerConfigPath          string
	genesisConfigPath       string
	memberSRVCConfigPath    string
	pbftConfigPath          string
	syncReplicaInfoInterval string
	syncReplica             bool
	license                 string
	rateLimitEnable         bool
	txRatePeak              int64
	txFillRate              string
	contractRatePeak        int64
	contractFillRate        string
	blockVersion            string
	transactionVersion      string
}

//return a config instances
func newconfigsImpl(globalConfigPath string, NodeID int, GRPCPort int, HTTPPort int, RESTPort int) *configsImpl {
	var cimpl configsImpl
	config := viper.New()
	viper.SetEnvPrefix("GLOBAL_ENV")
	config.SetConfigFile(globalConfigPath)
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error envPre %s reading %s", "GLOBAL", err))
	}
	/*
		system config
	 */
	cimpl.nodeID = NodeID
	cimpl.gRPCPort = GRPCPort
	cimpl.httpPort = HTTPPort
	cimpl.restPort = RESTPort
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

	/*
		Version
	 */
	cimpl.blockVersion = config.GetString("global.version.blockversion")
	cimpl.transactionVersion = config.GetString("global.version.transactionversion")
	return &cimpl
}

func (cIml *configsImpl) getNodeID() int            { return cIml.nodeID }
func (cIml *configsImpl) getGRPCPort() int          { return cIml.gRPCPort }
func (cIml *configsImpl) getHTTPPort() int          { return cIml.httpPort }
func (cIml *configsImpl) getRESTPort() int          { return cIml.restPort }
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
func (cIml *configsImpl) getRateLimitConfig () jsonrpc.RateLimitConfig {
	txFillRate, _ := time.ParseDuration(cIml.txFillRate)
	contractFillRate, _ := time.ParseDuration(cIml.contractFillRate)
	return jsonrpc.RateLimitConfig{
		Enable: cIml.rateLimitEnable,
		TxRatePeak: cIml.txRatePeak,
		TxFillRate: txFillRate,
		ContractRatePeak: cIml.contractRatePeak,
		ContractFillRate: contractFillRate,
	}
}
func (cIml *configsImpl) getBlockVersion() string {
	return cIml.blockVersion
}
func (cIml *configsImpl) getTransactionVersion() string {
	return cIml.transactionVersion
}
