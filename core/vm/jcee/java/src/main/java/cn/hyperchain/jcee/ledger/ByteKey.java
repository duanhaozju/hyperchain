package cn.hyperchain.jcee.ledger;

import java.util.Arrays;

/**
 * Created by huhu on 2017/6/22.
 */
public class ByteKey {
    private final byte[] key;

    public ByteKey(byte[] key) {
        this.key = key; // You may want to do a defensive copy here
    }

    public byte[] getKey() {
        return key;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        ByteKey cacheKey = (ByteKey) o;
        return Arrays.equals(key, cacheKey.key);
    }

    @Override
    public int hashCode() {
        return key != null ? Arrays.hashCode(key) : 0;
    }
}