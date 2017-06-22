package cn.hyperchain.jcee.mock;

import cn.hyperchain.jcee.ledger.*;
import cn.hyperchain.jcee.util.Bytes;
import cn.hyperchain.protos.ContractProto;
import com.google.protobuf.ByteString;

import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Created by huhu on 2017/6/21.
 */
public class MockLedger extends AbstractLedger {
    private Cache cache;
    public MockLedger(){
        cache = new HyperCache();
    }
    @Override
    public Batch newBatch() {
        return new BatchImpl(this);
    }

    class BatchImpl implements Batch{
        private Map<ByteKey, Result> data;
        private MockLedger ledger;

        public BatchImpl(MockLedger ledger) {
            data = new ConcurrentHashMap<>();
            this.ledger = ledger;
        }

        public Result get(byte[] key) {
            Result result = this.data.get(new ByteKey(key));
            if(result == null){
                return new Result(ByteString.EMPTY);
            }
            return result;
        }

        @Override
        public Result get(String key){
            return get(key.getBytes());
        }


        @Override
        public void put(byte[] key, byte[] value) {
            if(value == null || value.length == 0 ){
                data.put(new ByteKey(key),new Result(ByteString.EMPTY));
            }else {
                data.put(new ByteKey(key), new Result(ByteString.copyFrom(value)));
            }
        }

        @Override
        public void put(byte[] key, Object value) {
            put(key, Bytes.toByteArray(value));
        }

        @Override
        public void put(String key, Object value){
            put(key.getBytes(), Bytes.toByteArray(value));
        }

        @Override
        public void put(String key, byte[] value){
            put(key.getBytes(), value);
        }

        @Override
        public void reset() {
            data.clear();
        }

        @Override
        public boolean commit() {
            return ledger.writeBatch(this.toBatchKV());
        }

        public ContractProto.BatchKV toBatchKV() {
            ContractProto.BatchKV.Builder builder  = ContractProto.BatchKV.newBuilder();
            for(Map.Entry<ByteKey, Result> kv: data.entrySet()) {
                ContractProto.KeyValue keyValue = ContractProto.KeyValue.newBuilder()
                        .setK(ByteString.copyFrom(kv.getKey().getKey()))
                        .setV(kv.getValue().getValue())
                        .build();
                builder.addKv(keyValue);
            }
            return builder.build();
        }
    }

    public boolean writeBatch(ContractProto.BatchKV batch) {

        int count = batch.getKvCount();
        for(int i =0;i<count;i++){
            ContractProto.KeyValue data = batch.getKv(i);
            byte[] key = data.getK().toByteArray();
            byte[] value = data.getV().toByteArray();

            cache.put(key, value);
        }

        return true;
    }

    @Override
    public BatchKey newBatchKey() {
        return new BatchKeyImpl();
    }

    class BatchKeyImpl implements BatchKey {
        ContractProto.BatchKey.Builder builder;
        List<byte[]> keys;

        public BatchKeyImpl() {
            keys = new LinkedList<>();//TODO: remove the duplicate keys
            builder = ContractProto.BatchKey.newBuilder();
        }

        @Override
        public void put(byte[] key) {
            keys.add(key);
            builder.addK(ByteString.copyFrom(key));
        }

        public void put(String key){
            put(key.getBytes());
        }

        @Override
        public List<byte[]> getKeys() {
            return this.keys;
        }
    }

    @Override
    public Batch batchRead(BatchKey key) {
        Batch batch = this.newBatch();
        List<byte[]> keys = key.getKeys();

        for(byte[] k: keys){

            byte[] value = cache.get(k);
            batch.put(k,value);
        }
        return batch;
    }

    @Override
    public BatchValue rangeQuery(byte[] start, byte[] end) {
        return null;
    }

    @Override
    public boolean delete(byte[] key) {
        cache.delete(key);
        return true;
    }

    @Override
    public boolean delete(String key) {
        return delete(key.getBytes());
    }

    @Override
    public Result get(byte[] key) {

        byte[] data = cache.get(key);
        if(data != null){
            return new Result(ByteString.copyFrom(data));
        }

        return new Result(ByteString.EMPTY);
    }

    @Override
    public Result get(String key){
        return get(key.getBytes());
    }

    @Override
    public boolean put(byte[] key, byte[] value) {

        cache.put(key,value);
        return true;
    }

    @Override
    public boolean put(byte[] key, boolean value) {
        return put(key, Boolean.toString(value).getBytes());
    }

    @Override
    public boolean put(byte[] key, short value) {
        return put(key, Short.toString(value).getBytes());
    }

    @Override
    public boolean put(byte[] key, char value) {
        return put(key, Character.toString(value).getBytes());
    }

    @Override
    public boolean put(byte[] key, int value) {
        return put(key, Integer.toString(value).getBytes());
    }

    @Override
    public boolean put(byte[] key, float value) {
        return put(key, Float.toString(value).getBytes());
    }

    @Override
    public boolean put(byte[] key, double value) {
        return put(key, Double.toString(value).getBytes());
    }

    @Override
    public boolean put(byte[] key, String value) {
        return put(key, value.getBytes());
    }

    @Override
    public boolean put(byte[] key, Object object) {
        return put(key, Bytes.toByteArray(object));
    }

    @Override
    public boolean put(String key, byte[] value) {
        return put(key.getBytes(), value);
    }

    @Override
    public boolean put(String key, boolean value) {
        return put(key.getBytes(), value);
    }

    @Override
    public boolean put(String key, short value) {
        return put(key.getBytes(), value);
    }

    @Override
    public boolean put(String key, char value) {
        return put(key.getBytes(), value);
    }

    @Override
    public boolean put(String key, int value) {
        return put(key.getBytes(), value);
    }

    @Override
    public boolean put(String key, float value) {
        return put(key.getBytes(), value);
    }

    @Override
    public boolean put(String key, double value) {
        return put(key.getBytes(), value);
    }

    @Override
    public boolean put(String key, String value) {
        return put(key.getBytes(), value);
    }

    @Override
    public boolean put(String key, Object object) {
        return put(key.getBytes(), object);
    }
}
