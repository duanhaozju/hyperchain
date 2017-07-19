package cn.hyperchain.jcee.contract.filter;

import org.apache.log4j.Logger;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Created by wangxiaoyi on 2017/7/17.
 */
public class FilterChainMgr {

    private static final Logger LOG = Logger.getLogger(FilterChainMgr.class.getSimpleName());

    private Map<String, FilterChain> chainMap; // method to FilterChain map

    public FilterChainMgr() {
        chainMap = new ConcurrentHashMap<>();
    }

    /**
     * fetch filter chain for specified method
     * @param methodName
     * @return
     */
    public FilterChain getFilterChain(String methodName) {
        return chainMap.get(methodName);
    }

    public void AddFilterChain(String methodName, FilterChain chain) {
        chainMap.put(methodName, chain);
    }

    public void AddFilter(String methodName, Filter filter) {
        FilterChain fc = chainMap.get(methodName);
        if (fc == null) {
            fc = new FilterChain();
            chainMap.put(methodName, fc);
        }
        fc.addFilter(filter);
    }

    public void removeFilter(String methodName, String filterName) {
        FilterChain fc = chainMap.get(methodName);
        if (fc == null) {
            LOG.warn(String.format("no FilterChain found for method %s", methodName));
        }else {
            fc.removeFilter(filterName);
        }
    }

    public Filter getFilter(String methodName, String filterName) {
        FilterChain fc =chainMap.get(methodName);
        if (fc == null) {
            LOG.warn(String.format("no FilterChain found for method %s", methodName));
            return null;
        }else {
            return fc.getFilter(filterName);
        }
    }
}
