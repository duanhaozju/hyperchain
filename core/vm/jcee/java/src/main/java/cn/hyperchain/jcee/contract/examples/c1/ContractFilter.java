package cn.hyperchain.jcee.contract.examples.c1;

import cn.hyperchain.jcee.contract.filter.Filter;
import cn.hyperchain.jcee.executor.Context;

import java.util.HashSet;
import java.util.Set;

/**
 * Created by wangxiaoyi on 2017/7/21.
 */
class ContractFilter implements Filter {

    private Set<String> permittedContractAddrs;

    protected void addPermittedContractAddr(String contractAddr) {
        permittedContractAddrs.add(contractAddr);
    }

    public ContractFilter() {
        permittedContractAddrs = new HashSet<>();
    }

    @Override
    public boolean doFilter(Context context) {
        if (permittedContractAddrs.contains(context.getRequestContext().getCid())) {
            return true;
        }
        return false;
    }

    @Override
    public String getName() {
        return "ContractFilter";
    }
}