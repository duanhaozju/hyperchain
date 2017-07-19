package cn.hyperchain.jcee.contract.filter;

import cn.hyperchain.jcee.executor.Context;

/**
 * Created by huhu on 2017/7/3.
 */
public interface Filter {
    boolean doFilter(Context context);
    String getName();
}
