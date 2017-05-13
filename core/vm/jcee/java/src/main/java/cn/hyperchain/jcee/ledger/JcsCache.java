package cn.hyperchain.jcee.ledger;

import org.apache.commons.jcs.JCS;
import org.apache.commons.jcs.access.CacheAccess;
import org.apache.commons.jcs.access.exception.CacheException;
import org.apache.log4j.Logger;

/**
 * Created by huhu on 2017/5/11.
 */
public class JcsCache implements Cache{

    private CacheAccess<byte[], byte[]> cache = null;
    private static final Logger logger = Logger.getLogger(HyperchainLedger.class.getSimpleName());


    public JcsCache(){
        try {
            cache = JCS.getInstance("default");
        }
        catch (CacheException e){
            logger.info(String.format( "Problem initializing cache: %s", e.getMessage()));
        }
    }

    public void putInCache(byte[]key, byte[]value){
        try{
            cache.put( key, value);
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
