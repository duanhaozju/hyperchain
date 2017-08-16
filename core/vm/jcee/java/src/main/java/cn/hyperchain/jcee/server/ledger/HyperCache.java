/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.ledger;

import com.google.protobuf.ByteString;
import org.apache.commons.jcs.JCS;
import org.apache.commons.jcs.access.CacheAccess;
import org.apache.commons.jcs.access.exception.CacheException;
import org.apache.log4j.Logger;

import java.io.*;
import java.util.Properties;

/**
 * Created by huhu on 2017/5/11.
 */
public class HyperCache implements Cache{

    private CacheAccess<ByteString, byte[]> cache = null;
    private static final Logger logger = Logger.getLogger(HyperchainLedger.class.getSimpleName());
    private static final boolean cacheSwitcher = false;


    public HyperCache(){ //TODO: fix cache key may mix problem !!!
        try {
            Properties properties = new Properties();
            properties.load(new FileInputStream("./hyperjvm/config/cache.properties"));
            JCS.setConfigProperties(properties);
            cache = JCS.getInstance("default");
        }
        catch (CacheException e){
            logger.info(String.format( "Problem initializing cache: %s", e.getMessage()));
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    public void put(byte[]key, byte[]value){
        try{
            logger.debug("put in cache "+new String(value));
            cache.put( ByteString.copyFrom(key), value);
            logger.debug("cache size after put:"+size());
        }
        catch ( CacheException e){
            logger.info(String.format( "Problem putting object in the cache, for key %s%n%s",key, e.getMessage()));
        }
    }

    public byte[] get(byte[] key) {
        if (cacheSwitcher) {
            // logger.info("try to get from cache for key "+new String(key));
            return cache.get(ByteString.copyFrom(key));
        } else {
            return null;
        }
    }

    public void delete(byte[] key) {
        cache.remove(ByteString.copyFrom(key));
    }

    public int size(){
        return cache.getCacheControl().getMemoryCache().getSize();
    }
}
