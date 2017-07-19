/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract.filter;
import cn.hyperchain.jcee.executor.Context;
import org.apache.log4j.Logger;

import java.util.Iterator;
import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * Created by huhu on 2017/7/3.
 */
public class FilterChain {

    private static final Logger LOG = Logger.getLogger(FilterChain.class.getSimpleName());
    private List<Filter> filters = new CopyOnWriteArrayList<>();

    public boolean doFilter(Context context){
        Iterator<Filter> it = filters.iterator();
        while (it.hasNext()) {
            Filter filter = it.next();
            if (filter.doFilter(context) == false) {
                LOG.warn(context + " is not passed filter with name " + filter.getName());
                return false;
            }
        }
        return true;
    }

    public void addFilter(Filter filter) {
        this.filters.add(filter);
    }

    //remove fitler by filter name
    public void removeFilter(String filterName) {
        if (filterName == null || filterName.isEmpty()) {
            LOG.error("filter name is null or empty");
            return;
        }
        Iterator<Filter> it = filters.iterator();
        while (it.hasNext()) {
            Filter filter = it.next();
            if (filter.getName().equals(filterName)) {
                it.remove();
                return;
            }
        }
    }
}
