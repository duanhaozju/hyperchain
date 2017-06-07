package cn.hyperchain.jcee.ledger;


/**
 * Created by wangxiaoyi on 2017/5/5.
 */
public interface Batch {
    //Put keyvalue into the batch
    void put(byte[] key, byte[]value);
    void put(String key, Object value);

    byte[] get(byte[] key);
    <T> T get(String key, Class<T> clazz);

    //Clear data in batch
    void reset();

    //Commit batch data into ledger
    boolean commit();
}
