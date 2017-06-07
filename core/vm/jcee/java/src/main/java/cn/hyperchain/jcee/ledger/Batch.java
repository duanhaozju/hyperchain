package cn.hyperchain.jcee.ledger;


/**
 * Created by wangxiaoyi on 2017/5/5.
 */
public interface Batch {
    //Put keyvalue into the batch
    void put(byte[] key, byte[]value);
    void put(String key, Object value);

    Result get(byte[] key);
    Result get(String key);

    //Clear data in batch
    void reset();

    //Commit batch data into ledger
    boolean commit();
}
