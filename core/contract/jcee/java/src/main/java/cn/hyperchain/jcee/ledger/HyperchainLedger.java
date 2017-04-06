package cn.hyperchain.jcee.ledger;

import cn.hyperchain.jcee.ContractGrpcServerImpl;

/**
 * HyperchainLedger is an implementation of ILedger
 * which store manipulate the data using remote hyperchain server.
 */

public class HyperchainLedger implements ILedger{

    private ContractGrpcServerImpl cgsi;

    public HyperchainLedger(ContractGrpcServerImpl csgi) {
        this.cgsi = csgi;
    }

    public byte[] get(byte[] key) {
        return new byte[0];
    }

    public boolean put(byte[] key, byte[] value) {
        return false;
    }
}
