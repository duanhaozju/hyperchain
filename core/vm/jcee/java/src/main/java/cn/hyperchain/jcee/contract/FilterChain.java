package cn.hyperchain.jcee.contract;


import java.util.HashMap;
import java.util.Map;

/**
 * Created by huhu on 2017/7/3.
 */
public class FilterChain {

    private HashMap<String,Filter> filters = new HashMap<>();

    public boolean doFilter(ContractInfo info){
        for(Map.Entry<String,Filter> entry : filters.entrySet()){
            if(!entry.getValue().doFilter(info)){
                return false;
            }
        }
        return true;
    }

    public HashMap<String,Filter> getFilters() {
        return filters;
    }
    public Filter getFilter(String key){
        return filters.get(key);
    }
    public void addFilter(String key,Filter f){
        filters.put(key,f);
    }
    public void removeFilter(String key){
        filters.remove(key);
    }
}
