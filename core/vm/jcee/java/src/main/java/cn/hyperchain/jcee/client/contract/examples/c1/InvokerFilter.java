package cn.hyperchain.jcee.client.contract.examples.c1;

import cn.hyperchain.jcee.client.contract.filter.Filter;
import cn.hyperchain.jcee.common.Context;

import java.util.HashSet;
import java.util.Set;

/**
 * Created by wangxiaoyi on 2017/7/21.
 */
class InvokerFilter implements Filter {

    private Set<String> permittedInvokerAddrs;

    protected void addPermittedInvokerAddr(String invokerAddr) {
        permittedInvokerAddrs.add(invokerAddr);
    }

    public InvokerFilter() {
        permittedInvokerAddrs = new HashSet<>();
    }

    @Override
    public boolean doFilter(Context context) {
        if (permittedInvokerAddrs.contains(context.getRequestContext().getInvoker())) {
            return true;
        }
        return false;
    }

    @Override
    public String getName() {
        return "InvokerFilter";
    }
}