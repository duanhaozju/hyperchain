/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger;

import cn.hyperchain.protos.ContractProto;
import com.google.protobuf.ByteString;
import org.apache.log4j.Logger;

/**
 * HyperchainLedger is an implementation of AbstractLedger
 * which store manipulate the data using remote hyperchain server.
 */
public class HyperchainLedger extends AbstractLedger{

    private static final Logger logger = Logger.getLogger(HyperchainLedger.class.getSimpleName());
    private LedgerClient client;
    public HyperchainLedger(int port){
        client = new LedgerClient("localhost", port);
    }

    public byte[] get(byte[] key) {
        ContractProto.Key sendkey = ContractProto.Key.newBuilder()
                .setContext(getLedgerContext())
                .setK(ByteString.copyFrom(key))
                .build();
        logger.info("Transaction id: " + getContext().getId());
        return client.get(sendkey).getV().toByteArray();
    }

    public boolean put(byte[] key, byte[] value) {
        ContractProto.KeyValue kv = ContractProto.KeyValue.newBuilder()
                .setContext(getLedgerContext())
                .setK(ByteString.copyFrom(key))
                .setV(ByteString.copyFrom(value))
                .build();
        return client.put(kv);
    }

    public ContractProto.Value fetch(byte[] key) {
        ContractProto.Key sendkey = ContractProto.Key.newBuilder()
                .setContext(getLedgerContext())
                .setK(ByteString.copyFrom(key))
                .build();
        logger.info("Transaction id: " + getContext().getId());
        return client.get(sendkey);
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
}
