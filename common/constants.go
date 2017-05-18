//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

//constants for logger key
const (
	LOG_FILE_DIR          = "global.logs.logsdir"
	LOG_DUMP_FILE 	      = "global.logs.dumpfile"
	LOG_BASE_LOG_LEVEL    = "global.logs.loglevel"
	LOG_NEW_FILE_INTERVAL = "global.logs.newLogFileInterval"
	LOG_FILE_FORMAT       = "global.logs.file_format"
	LOG_CONSOLE_FORMAT    = "global.logs.console_format"
	LOG_MODULE_KEY        = "global.logs.module"
	DEFAULT_LOG        = "system"
)

//constants for input args
const (
	C_NODE_ID            = "global.id"
	C_NODE_IP            = "global.ip"
	C_GRPC_PORT          = "global.grpc_port"
	C_HTTP_PORT          = "global.http_port"
	C_REST_PORT          = "global.rest_port"
	C_WEBSOCKET_PORT     = "global.websocket_port"
	C_PEER_CONFIG_PATH   = "global.peerconfigs.path"
	C_GLOBAL_CONFIG_PATH = "global.globalconfig.path"
)

const (

	NAMESPACE = "namespace.name"
	DEFAULT_NAMESPACE = "system"
	KEY_STORE_DIR = "global.account.keystoredir"
	KEY_NODE_DIR  = "global.account.keynodesdir"
	START_NAMESPACE = "global.namespace.start."

	LOG_DUMP_FILE_FLAG = "global.logs.dumpfile"
	LOG_DUMP_FILE_DIR  = "global.logs.logsdir"
	LOG_LEVEL          = "global.logs.loglevel"

	DATABASE_DIR = "global.database.dir"

	PEER_CONFIG_PATH = "global.configs.peers"

	GENESIS_CONFIG_PATH = "global.configs.genesis"

	MEMBER_SRVC_CONFIG_PATH = "global.configs.membersrvc"

	PBFT_CONFIG_PATH = "global.configs.pbft"

	DB_CONFIG_PATH             = "global.dbConfig"
	SYNC_REPLICA_INFO_INTERVAL = "global.configs.replicainfo.interval"

	SYNC_REPLICA = "global.configs.replicainfo.enable"
	LICENSE      = "global.configs.license"

	RATE_LIMIT_ENABLE  = "global.configs.ratelimit.enable"
	TX_RATE_PEAK       = "global.configs.ratelimit.txRatePeak"
	TX_FILL_RATE       = "global.configs.ratelimit.txFillRate"
	CONTRACT_RATE_PEAK = "global.configs.ratelimit.contractRatePeak"
	CONTRACT_FILL_RATE = "global.configs.ratelimit.contractFillRate"
	STATE_TYPE         = "global.structure.state"

	BLOCK_VERSION       = "global.version.blockversion"
	TRANSACTION_VERSION = "global.version.transactionversion"

	PAILL_PUBLIC_KEY_N       = "global.configs.hmpublickey.N"
	PAILL_PUBLIC_KEY_NSQUARE = "global.configs.hmpublickey.Nsquare"

	PAILL_PUBLIC_G = "global.configs.hmpublickey.G"

	//Bucket tree
	STATE_SIZE          = "global.configs.buckettree.state.size"
	STATE_LEVEL_GROUP   = "global.configs.buckettree.state.levelGroup"
	STORAGE_SIZE        = "global.configs.buckettree.storage.size"
	STORAGE_LEVEL_GROUP = "global.configs.buckettree.storage.levelGroup"
)
