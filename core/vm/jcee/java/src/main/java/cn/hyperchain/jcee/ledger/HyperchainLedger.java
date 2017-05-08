/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger;

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
    public HyperchainLedger(int port){
        ledgerClient = new LedgerClient("localhost", port);
    }

    public byte[] get(byte[] key) {
        ContractProto.Key sendkey = ContractProto.Key.newBuilder()
                .setContext(getLedgerContext())
                .setK(ByteString.copyFrom(key))
                .build();
        logger.info("Transaction id: " + getContext().getId());
        return ledgerClient.get(sendkey).getV().toByteArray();
    }

    public boolean put(byte[] key, byte[] value) {
        ContractProto.KeyValue kv = ContractProto.KeyValue.newBuilder()
                .setContext(getLedgerContext())
                .setK(ByteString.copyFrom(key))
                .setV(ByteString.copyFrom(value))
                .build();
        return ledgerClient.put(kv);
    }

    public ContractProto.Value fetch(byte[] key) {
        ContractProto.Key sendkey = ContractProto.Key.newBuilder()
                .setContext(getLedgerContext())
                .setK(ByteString.copyFrom(key))
                .build();
        logger.info("Transaction id: " + getContext().getId());
        return ledgerClient.get(sendkey);
    }

    public ContractProto.LedgerContext getLedgerContext(){
        return ContractProto.LedgerContext
                .newBuilder()
                .setNamespace(getContext().getRequestContext().getNamespace())
                .setTxid(getContext().getRequestContext().getTxid())
                .setCid(getContext().getRequestContext().getCid())
                .build();
    }

    public boolean batchRead(ContractProto.BatchKey key) {
        return false;
    }


    @Override
    public boolean getBoolean(byte[] key) {
        return Boolean.parseBoolean(fetch(key).getV().toStringUtf8());
    }

    @Override
    public short getShort(byte[] key) {
        return Short.parseShort(fetch(key).getV().toStringUtf8());
    }

    @Override
    public char getChar(byte[] key) {
        return fetch(key).getV().toStringUtf8().charAt(0);
    }

    @Override
    public int getInt(byte[] key) {
        return Integer.parseInt(fetch(key).getV().toStringUtf8());
    }

    @Override
    public float getFloat(byte[] key) {
        return Float.parseFloat(fetch(key).getV().toStringUtf8());
    }

    public double getDouble(byte[] key){
        ContractProto.Value v = fetch(key);
        return Double.parseDouble(v.getV().toStringUtf8());
    }

    @Override
    public String getString(byte[] key) {
        return fetch(key).getV().toStringUtf8();
    }

    @Override
    public Object getObject(byte[] key) {
        //TODO: get object, use model method
        return null;
    }

    @Override
    public byte[] get(String key) {
        return get(key.getBytes());
    }

    @Override
    public boolean getBoolean(String key) {
        return getBoolean(key.getBytes());
    }

    @Override
    public short getShort(String key) {
        return getShort(key.getBytes());
    }

    @Override
    public char getChar(String key) {
        return getChar(key.getBytes());
    }

    @Override
    public int getInt(String key) {
        return getInt(key.getBytes());
    }

    @Override
    public float getFloat(String key) {
        return getFloat(key.getBytes());
    }

    @Override
    public double getDouble(String key) {
        return getDouble(key.getBytes());
    }

    @Override
    public String getString(String key) {
        return getString(key.getBytes());
    }

    @Override
    public Object getObject(String key) {
        //TODO: get object, use model method
        return null;
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
//        return put(key);
        //TODO: encode the object
        return false;
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
        return ledgerClient.batchWrite(batch);
    }

    @Override
    public BatchKey newBatchKey() {
        return new BatchKeyImpl();
    }

    @Override
    public Batch batchRead(BatchKey key) {
        ContractProto.BathValue bv = ledgerClient.bathRead(toProtoBatchKey(key));
        Batch batch = this.newBatch();
        List<ByteString> values = bv.getVList();
        List<byte[]> keys = key.getKeys();
        int i = 0;
        for (byte[] k: keys) {
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
                        .setContext(getLedgerContext())
                        .build();
                builder.addKv(keyValue);
            }
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
