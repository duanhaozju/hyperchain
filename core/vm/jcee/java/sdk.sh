#!/usr/bin/env bash
function initClass() {
    contract="target/classes/cn/hyperchain/jcee/contract"
    sdkContract="sdk/cn/hyperchain/jcee/contract"

    ledger="target/classes/cn/hyperchain/jcee/ledger"
    sdkLedger="sdk/cn/hyperchain/jcee/ledger"

    common="target/classes/cn/hyperchain/jcee/common"
    sdkCommon="sdk/cn/hyperchain/jcee/common"

    util="target/classes/cn/hyperchain/jcee/util"
    sdkUtil="sdk/cn/hyperchain/jcee/util"

    log="target/classes/cn/hyperchain/jcee/log"
    sdkLog="sdk/cn/hyperchain/jcee/log"

    mkdir -p ${sdkContract}
    echo "${contract}/ContractTemplate.class"
    cp "${contract}/ContractTemplate.class" ${sdkContract}

    mkdir -p ${sdkLedger}
    cp "${ledger}/ILedger.class" ${sdkLedger}
    cp "${ledger}/AbstractLedger.class" ${sdkLedger}
    cp "${ledger}/Batch.class" ${sdkLedger}
    cp "${ledger}/BatchKey.class" ${sdkLedger}
    cp "${ledger}/BatchValue.class" ${sdkLedger}

    mkdir -p ${sdkCommon}
    cp "${common}/ExecuteResult.class" ${sdkCommon}

    mkdir -p "${sdkCommon}/exception"
    cp "${common}/exception/NotExistException.class" "${sdkCommon}/exception"
    cp "${common}/exception/HyperjvmException.class" "${sdkCommon}/exception"

    mkdir -p ${sdkUtil}
    cp "${util}/Bytes.class" ${sdkUtil}

    mkdir -p ${sdkLog}
    cp "${log}/Logger.class" ${sdkLog}
    cp "${log}/LoggerImpl.class" ${sdkLog}
}

if [ -e target ]; then
    echo "delete target."
    rm -rf target
fi

if [ -e sdk ]; then
    echo "delete sdk."
    rm -rf sdk
fi

echo "1. build target."
mvn clean package -Dmaven.test.skip=true

echo "2. create directory."
mkdir sdk

echo "3. paste needed class."
initClass

echo "4. build jar."
cd sdk
jar -cvf sdk.jar cn

