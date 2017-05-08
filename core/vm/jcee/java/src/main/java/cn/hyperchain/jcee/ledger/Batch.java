package cn.hyperchain.jcee.ledger;


/**
 * Created by wangxiaoyi on 2017/5/5.
 */
public interface Batch {
    //Put keyvalue into the batch
    void put(byte[] key, byte[]value);

    //Clear data in batch
    void reset();

    //Commit batch data into ledger
    boolean commit();
}
