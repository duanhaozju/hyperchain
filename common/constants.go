//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

//constants for input args
const (
	C_NODE_ID = "self.id"
)

const (
	OP_ARCHIVE = 100
)

const (
	NAMESPACE          = "namespace.name"
	DEFAULT_NAMESPACE  = "system"
	PEER_CONFIG_PATH   = "config.path.peerconfig"
	GLOBAL_CONFIG_PATH = "global.config.path"
)

const (
	START_NAMESPACE = "namespace.start."
	C_JVM_START     = "hypervm.jvm"
)

//constants for logger keys
const (
	LOG_DUMP_FILE_DIR     = "log.log_dir"
	LOG_DUMP_FILE         = "log.dump_file"
	LOG_NEW_FILE_INTERVAL = "log.dump_interval"
	LOG_BASE_LOG_LEVEL    = "log.log_level"
	LOG_FILE_FORMAT       = "log.file_format"
	LOG_CONSOLE_FORMAT    = "log.console_format"
	LOG_MODULE_KEY        = "log.module"
	DEFAULT_LOG           = "system"
	LOG_MAX_SIZE          = "log.max_log_size"
)

//constants for port keys
const (
	JSON_RPC_PORT  = "port.jsonrpc"
	JVM_PORT       = "port.jvm"
	LEDGER_PORT    = "port.ledger"
	P2P_PORT       = "port.grpc"
	WEBSOCKET_PORT = "port.websocket"
)

//constants for p2p configuration keys
const (
	P2P_RETRY_TIME               = "p2p.retrytime"
	P2P_IPC                      = "p2p.ipc"
	P2P_SERVERNAME               = "p2p.servername"
	P2P_ENABLE_TLS               = "p2p.enableTLS"
	P2P_TLS_CA                   = "p2p.tlsCA"
	P2P_TLS_SERVER_HOST_OVERRIDE = "p2p.tlsServerHostOverride"
	P2P_TLS_CERT                 = "p2p.tlsCert"
	P2P_TLS_CERT_PRIV            = "p2p.tlsCertPriv"
	P2P_HOSTS                    = "p2p.hosts"
	P2P_ADDR                     = "P2P.addr"
)

// constants for http configuration keys
const (
	HTTP_SECURITY       = "http.security"
	HTTP_VERSION2       = "http.http_2"
	HTTP_ALLOWEDORIGINS = "http.allowedOrigins"
)

//constants for encryption configuration keys
const (
	ENCRYPTION_ECERT_ECA   = "encryption.ecert.eca"
	ENCRYPTION_ECERT_ECERT = "encryption.ecert.ecert"
	ENCRYPTION_ECERT_PRIV  = "encryption.ecert.priv"

	ENCRYPTION_RCERT_RCA   = "encryption.rcert.rca"
	ENCRYPTION_RCERT_RCERT = "encryption.rcert.rcert"
	ENCRYPTION_RCERT_PRIV  = "encryption.rcert.priv"

	ENCRYPTION_TCERT_WHITELIST     = "encryption.tcert.whiteList"
	ENCRYPTION_TCERT_WHITELIST_DIR = "encryption.tcert.listDir"

	ENCRYPTION_CHECK_ENABLE   = "encryption.check.enable"
	ENCRYPTION_CHECK_SIGN     = "encryption.check.sign"
	ENCRYPTION_CHECK_ENABLE_T = "encryption.check.enableT"

	ENCRYPTION_SECURITY_ALGO = "encryption.security.algo"
)

//constants for administrator oprations
const (
	ADMIN_CHECK      = "admin.check"
	ADMIN_EXPIRATION = "admin.expiration"
)
