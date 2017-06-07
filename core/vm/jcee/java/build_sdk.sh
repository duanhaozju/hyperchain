#!/usr/bin/env bash
function initClass() {
    contract="target/classes/cn/hyperchain/jcee/contract"
    sdkContract="hyperjvm/sdk/cn/hyperchain/jcee/contract"

    ledger="target/classes/cn/hyperchain/jcee/ledger"
    sdkLedger="hyperjvm/sdk/cn/hyperchain/jcee/ledger"

    common="target/classes/cn/hyperchain/jcee/common"
    sdkCommon="hyperjvm/sdk/cn/hyperchain/jcee/common"

    util="target/classes/cn/hyperchain/jcee/util"
    sdkUtil="hyperjvm/sdk/cn/hyperchain/jcee/util"

    log="target/classes/cn/hyperchain/jcee/log"
    sdkLog="hyperjvm/sdk/cn/hyperchain/jcee/log"

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

if [ -e hyperjvm/sdk ]; then
    echo "delete sdk."
    rm -rf hyperjvm/sdk
fi

echo "3-1. create directory."
mkdir hyperjvm/sdk

echo "3-2. paste needed class."
initClass

echo "3-3. build jar."
cd hyperjvm/sdk
jar -cvf hyperjvm-sdk-1.0.jar cn

