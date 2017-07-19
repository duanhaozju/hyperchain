package cn.hyperchain.jcee.contract.filter;

import cn.hyperchain.jcee.executor.Context;
import org.junit.AfterClass;
import org.junit.BeforeClass;
import org.junit.Test;

import java.util.HashSet;
import java.util.Set;

/**
 * Created by wangxiaoyi on 2017/7/18.
 */
public class TestFilterChainMgr {

    static FilterChainMgr mgr;

    @BeforeClass
    public static void setup() {
        mgr = new FilterChainMgr();
    }

    @AfterClass
    public static void destroy() {

    }

    @Test
    public void testAddGetFilterChain() {
        FilterChain fc = new FilterChain();
        fc.addFilter(new NamespaceFilter());

    }

    class NamespaceFilter implements Filter{

        private Set<String> permittedNamespaces;

        protected void addPermittedNamespace(String namespace) {
            permittedNamespaces.add(namespace);
        }

        public NamespaceFilter() {
           permittedNamespaces = new HashSet<>();
           permittedNamespaces.add("ns1");
           permittedNamespaces.add("ns2");
        }

        @Override
        public boolean doFilter(Context context) {
            if (permittedNamespaces.contains(context.getRequestContext().getNamespace())) {
                return true;
            }
            return false;
        }

        @Override
        public String getName() {
            return "NamespaceFilter";
        }
    }


    class ContractFilter implements Filter{

        private Set<String> permittedContractAddrs;

        protected void setPermittedContractAddrs(String contractAddr) {
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


    class InvokerFilter implements Filter {

        private Set<String> permittedInvokerAddrs;

        protected void setPermittedInvokerAddrs(String invokerAddr) {
            permittedInvokerAddrs.add(invokerAddr);
        }

        public InvokerFilter() {
            permittedInvokerAddrs = new HashSet<>();
            permittedInvokerAddrs.add("xxx");
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
}
