/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger;

import cn.hyperchain.jcee.common.exception.NotExistException;
import cn.hyperchain.jcee.util.Bytes;
import cn.hyperchain.protos.ContractProto;
import com.google.protobuf.ByteString;
import org.apache.log4j.Logger;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

/**
 * HyperchainLedger is an implementation of AbstractLedger
 * which store manipulate the data using remote hyperchain server.
 */
public class HyperchainLedger extends AbstractLedger{

    private static final Logger logger = Logger.getLogger(HyperchainLedger.class.getSimpleName());
    private LedgerClient ledgerClient;
    private Cache cache;
    public HyperchainLedger(int port){
        ledgerClient = new LedgerClient("localhost", port);
        cache = new JcsCache();
    }

    public byte[] get(byte[] key) throws NotExistException {
        byte[] data = cache.retrieveFromCache(key);
        if(data != null){
            return data;
        }

        ContractProto.Key sendkey = ContractProto.Key.newBuilder()
                .setContext(getLedgerContext())
                .setK(ByteString.copyFrom(key))
                .build();
        logger.debug("Transaction id: " + getContext().getId());

        ByteString v = ledgerClient.get(sendkey).getV();

        if (v == null || v.isEmpty()){
            throw new NotExistException(new String(key));
        }
        cache.putInCache(key,v.toByteArray());
        return v.toByteArray();
    }

    public boolean put(byte[] key, byte[] value) {
        ContractProto.KeyValue kv = ContractProto.KeyValue.newBuilder()
                .setContext(getLedgerContext())
                .setK(ByteString.copyFrom(key))
                .setV(ByteString.copyFrom(value))
                .build();

        logger.info("the value put in ledger "+ByteString.copyFrom(value));
        boolean success = ledgerClient.put(kv);
        if(success){
            logger.info("the value put in cache "+value.toString());
            cache.putInCache(key,value);
        }
        return success;
    }

    public ContractProto.Value fetch(byte[] key) {
        byte[] data = cache.retrieveFromCache(key);
        if(data!=null){
            ContractProto.Value recvValue = ContractProto.Value.newBuilder()
                    .setV(ByteString.copyFrom(data))
                    .build();
            return recvValue;
        }
        ContractProto.Key sendkey = ContractProto.Key.newBuilder()
                .setContext(getLedgerContext())
                .setK(ByteString.copyFrom(key))
                .build();
        logger.info("Transaction id: " + getContext().getId());

        ContractProto.Value value = ledgerClient.get(sendkey);
        cache.putInCache(key,value.toByteArray());
        return value;
    }

    public ContractProto.LedgerContext getLedgerContext(){
        return ContractProto.LedgerContext
                .newBuilder()
                .setNamespace(getContext().getRequestContext().getNamespace())
                .setTxid(getContext().getRequestContext().getTxid())
                .setCid(getContext().getRequestContext().getCid())
                .build();
    }

    @Override
    public boolean delete(byte[] key) {
        ContractProto.Key ck = ContractProto.Key.newBuilder()
                .setContext(getLedgerContext())
                .setK(ByteString.copyFrom(key))
                .build();
        boolean success = ledgerClient.delete(ck);
        if(success){
            cache.removeFromCache(key);
        }
        return success;
    }

    @Override
    public boolean delete(String key) {
        return delete(key.getBytes());
    }

    @Override
    public boolean getBoolean(byte[] key) throws NotExistException {
        ContractProto.Value v = fetch(key);
        if (v.getV().isEmpty())
            throw new NotExistException(new String(key));
        return Boolean.parseBoolean(v.getV().toStringUtf8());
    }

    @Override
    public short getShort(byte[] key) throws NotExistException {
        ContractProto.Value v = fetch(key);
        if (v.getV().isEmpty())
            throw new NotExistException(new String(key));
        return Short.parseShort(v.getV().toStringUtf8());
    }

    @Override
    public char getChar(byte[] key) throws NotExistException {
        ContractProto.Value v = fetch(key);
        if (v.getV().isEmpty())
            throw new NotExistException(new String(key));
        return v.getV().toStringUtf8().charAt(0);
    }

    @Override
    public int getInt(byte[] key) throws NotExistException {
        ContractProto.Value v = fetch(key);
        if (v.getV().isEmpty())
            throw new NotExistException(new String(key));
        return Integer.parseInt(v.getV().toStringUtf8());
    }

    @Override
    public float getFloat(byte[] key) throws NotExistException {
        ContractProto.Value v = fetch(key);
        if (v.getV().isEmpty())
            throw new NotExistException(new String(key));
        return Float.parseFloat(v.getV().toStringUtf8());
    }

    public double getDouble(byte[] key) throws NotExistException {
        ContractProto.Value v = fetch(key);
        if (v.getV().isEmpty())
            throw new NotExistException(new String(key));
        return Double.parseDouble(v.getV().toStringUtf8());
    }

    @Override
    public String getString(byte[] key) throws NotExistException {
        ContractProto.Value v = fetch(key);
        if (v.getV().isEmpty())
            throw new NotExistException(new String(key));
        return v.getV().toStringUtf8();
    }

    public  <T> T getObject(byte[] key, Class<T> clazz) {
        byte[] data = get(key);
        if(data == null){
            return null;
        }
        return Bytes.toObject(data, clazz);
    }

    @Override
    public byte[] get(String key) throws NotExistException {
        return get(key.getBytes());
    }

    @Override
    public boolean getBoolean(String key) throws NotExistException {
        return getBoolean(key.getBytes());
    }

    @Override
    public short getShort(String key) throws NotExistException {
        return getShort(key.getBytes());
    }

    @Override
    public char getChar(String key) throws NotExistException {
        return getChar(key.getBytes());
    }

    @Override
    public int getInt(String key) throws NotExistException {
        return getInt(key.getBytes());
    }

    @Override
    public float getFloat(String key) throws NotExistException {
        return getFloat(key.getBytes());
    }

    @Override
    public double getDouble(String key) throws NotExistException {
        return getDouble(key.getBytes());
    }

    @Override
    public String getString(String key) throws NotExistException {
        return getString(key.getBytes());
    }

    public  <T> T getObject(String key, Class<T> clazz) {
        byte[] data = get(key);
        if(data == null){
            return null;
        }
        return Bytes.toObject(data, clazz);
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

    @Override
    public Batch newBatch() {
        return new BatchImpl(this);
    }

    public boolean writeBatch(ContractProto.BatchKV batch) {

        boolean success = ledgerClient.batchWrite(batch);
        if(success){
            int count = batch.getKvCount();
            for(int i =0;i<count;i++){
                ContractProto.KeyValue data = batch.getKv(i);
                byte[] key = data.getK().toByteArray();
                byte[] value = data.getV().toByteArray();
                cache.putInCache(key, value);
            }
        }

        return success;
    }

    @Override
    public BatchKey newBatchKey() {
        return new BatchKeyImpl();
    }

    @Override
    public Batch batchRead(BatchKey key) {
        Batch batch = this.newBatch();
        List<byte[]> keys = key.getKeys();
        BatchKey bk = newBatchKey();

        for(byte[] k: keys){
            byte[] value = cache.retrieveFromCache(k);
            if(value!=null){
                batch.put(k,value);
            }
            else {
                bk.put(k);
            }
        }
        ContractProto.BathValue bv = ledgerClient.bathRead(toProtoBatchKey(bk));
        List<ByteString> values = bv.getVList();
        int i = 0;
        for (byte[] k: bk.getKeys()) {
            batch.put(k, values.get(i).toByteArray());
            i ++;
        }
        return batch;
    }

    @Override
    public BatchValue rangeQuery(byte[] start, byte[] end) {
        ContractProto.Range range = ContractProto.Range.newBuilder()
                .setStart(ByteString.copyFrom(start))
                .setEnd(ByteString.copyFrom(end))
                .setContext(getLedgerContext())
                .build();
        return new BathValueImpl(ledgerClient.rangeQuery(range));
    }

    class BatchImpl implements Batch{
        private Map<byte[], byte[]> data;
        private HyperchainLedger ledger;

        public BatchImpl(HyperchainLedger ledger) {
            data = new ConcurrentHashMap<>();
            this.ledger = ledger;
        }

        public byte[] get(byte[] key) {
            return this.data.get(key);
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
        public boolean commit() {
            return ledger.writeBatch(this.toBatchKV());
        }

        public ContractProto.BatchKV toBatchKV() {
            ContractProto.BatchKV.Builder builder  = ContractProto.BatchKV.newBuilder();
            for(Map.Entry<byte[], byte[]> kv: data.entrySet()) {
                ContractProto.KeyValue keyValue = ContractProto.KeyValue.newBuilder()
                        .setK(ByteString.copyFrom(kv.getKey()))
                        .setV(ByteString.copyFrom(kv.getValue()))
                        .build();
                builder.addKv(keyValue);
            }
            builder.setContext(getLedgerContext());
            return builder.build();
        }
    }

    private ContractProto.BatchKey toProtoBatchKey(BatchKey key) {
        ContractProto.BatchKey cbk = null;
        if (key instanceof BatchKeyImpl) {
            cbk = ((BatchKeyImpl) key).builder
                    .setContext(getLedgerContext())
                    .build();
        }
       return cbk;
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

        @Override
        public List<byte[]> getKeys() {
            return this.keys;
        }
    }

    class BathValueImpl implements BatchValue {

        Iterator<ContractProto.BathValue> rangeBatchValue;
        Iterator<ByteString> currBatchValue;

        public BathValueImpl(Iterator<ContractProto.BathValue> rangeBatchValue){
            this.rangeBatchValue = rangeBatchValue;
        }

        @Override
        public byte[] next() {
            if (hasNext()) {
                return currBatchValue.next().toByteArray();
            }else {
                throw new NoSuchElementException("No more value to display");
            }
        }

        @Override
        public boolean hasNext() {
            if (currBatchValue == null) {
                currBatchValue = rangeBatchValue.next().getVList().iterator();
            }
            if (currBatchValue.hasNext()) return true;
            if (rangeBatchValue.hasNext()) {
                currBatchValue = rangeBatchValue.next().getVList().iterator();
                return currBatchValue.hasNext();
            }
            return false;
        }
    }

}
