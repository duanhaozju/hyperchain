/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger;

import org.apache.commons.jcs.JCS;
import org.apache.commons.jcs.access.CacheAccess;
import org.apache.commons.jcs.access.exception.CacheException;
import org.apache.log4j.Logger;

import java.io.*;
import java.util.Properties;

/**
 * Created by huhu on 2017/5/11.
 */
public class JcsCache implements Cache{

    private CacheAccess<byte[], byte[]> cache = null;
    private static final Logger logger = Logger.getLogger(HyperchainLedger.class.getSimpleName());


    public JcsCache(){
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

    public void putInCache(byte[]key, byte[]value){
        try{
            logger.debug("put in cache "+new String(value));
            cache.put( key, value);
            logger.debug("cache size after put:"+size());
        }
        catch ( CacheException e){
            logger.info(String.format( "Problem putting object in the cache, for key %s%n%s",key, e.getMessage()));
        }
    }

    public byte[] retrieveFromCache(byte[] key)
    {
        return cache.get(key);
    }

    public void removeFromCache(byte[] key){
        cache.remove(key);
    }

    public int size(){
        return cache.getCacheControl().getMemoryCache().getSize();
    }
}
