package cn.hyperchain.jcee.ledger;

import java.util.List;

/**
 * Created by wangxiaoyi on 2017/5/8.
 */
public interface BatchKey {
    void put(byte[] key);
    void put(String key);
    List<byte[]> getKeys();
}
