package cn.hyperchain.jcee.ledger;

import cn.hyperchain.protos.ContractProto;
import com.google.protobuf.ByteString;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Created by wangxiaoyi on 2017/5/5.
 */
public class BatchImpl implements Batch{

    Map<byte[], byte[]> data;
    public BatchImpl() {
        data = new ConcurrentHashMap<>();
    }

    @Override
    public void put(byte[] key, byte[] value) {
        data.put(key, value);
    }

    @Override
    public void reset() {
        data.clear();
    }

    @Override
    public ContractProto.BatchKV toBatchKV() {
        ContractProto.BatchKV.Builder builder  = ContractProto.BatchKV.newBuilder();
        for(Map.Entry<byte[], byte[]> kv: data.entrySet()) {
            ContractProto.KeyValue keyValue = ContractProto.KeyValue.newBuilder()
                    .setK(ByteString.copyFrom(kv.getKey()))
                    .setV(ByteString.copyFrom(kv.getValue()))
                    .build();
            builder.addKv(keyValue);
        }
        return builder.build();
    }
}
