package cn.hyperchain.jcee.ledger;

import cn.hyperchain.protos.ContractProto;

/**
 * Created by wangxiaoyi on 2017/5/5.
 */
public interface Batch {
    void put(byte[] key, byte[]value);
    void reset();
    ContractProto.BatchKV toBatchKV();
}
