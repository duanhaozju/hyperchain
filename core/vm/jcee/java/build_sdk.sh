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

    mock="target/classes/cn/hyperchain/jcee/mock"
    sdkMock="hyperjvm/sdk/cn/hyperchain/jcee/mock"

    mkdir -p ${sdkContract}
    echo "${contract}/ContractTemplate.class"
    cp "${contract}/ContractTemplate.class" ${sdkContract}
    cp "${contract}/Event.class" ${sdkContract}

    mkdir -p ${sdkLedger}
    cp "${ledger}/ILedger.class" ${sdkLedger}
    cp "${ledger}/AbstractLedger.class" ${sdkLedger}
    cp "${ledger}/Batch.class" ${sdkLedger}
    cp "${ledger}/BatchKey.class" ${sdkLedger}
    cp "${ledger}/BatchValue.class" ${sdkLedger}
    cp "${ledger}/Result.class" ${sdkLedger}

    mkdir -p ${sdkCommon}
    cp "${common}/ExecuteResult.class" ${sdkCommon}

    mkdir -p ${sdkUtil}
    cp "${util}/Bytes.class" ${sdkUtil}

    mkdir -p ${sdkMock}
    cp "${mock}/MockLedger.class" ${sdkMock}
    cp "${mock}/MockServer.class" ${sdkMock}
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
cp hyperjvm/libs/log4j-1.2.17.jar hyperjvm/sdk
cd hyperjvm/sdk
jar -xvf log4j-1.2.17.jar
jar -cvf hyperjvm-sdk-1.0.jar cn org

echo "3-4. rm class."
rm -rf cn org META-INF
rm log4j-1.2.17.jar