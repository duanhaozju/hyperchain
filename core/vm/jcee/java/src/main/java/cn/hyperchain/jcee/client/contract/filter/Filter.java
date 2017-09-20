package cn.hyperchain.jcee.client.contract.filter;

import cn.hyperchain.jcee.common.Context;

/**
 * Created by huhu on 2017/7/3.
 */
public interface Filter {
    boolean doFilter(Context context);
    String getName();
}
