package cn.hyperchain.jcee.contract.filter;

import cn.hyperchain.jcee.executor.Context;
import cn.hyperchain.protos.ContractProto;
import org.junit.AfterClass;
import org.junit.Assert;
import org.junit.BeforeClass;
import org.junit.Test;

import java.util.HashSet;
import java.util.Set;

/**
 * Created by wangxiaoyi on 2017/7/18.
 */
public class TestFilterChainMgr {

    static FilterManager mgr;

    @BeforeClass
    public static void setup() {
        mgr = new FilterManager();
    }

    @AfterClass
    public static void destroy() {

    }

    @Test
    public void testFilter() {

        mgr.AddFilter("method1", new NamespaceFilter());
        Context context = new Context("001");
        context.setRequestContext(ContractProto
                .RequestContext.newBuilder()
                .setNamespace("ns1")
                .build());
        Assert.assertEquals(true, mgr.getFilterChain("method1").doFilter(context));

        Context context1 = new Context("001");
        context1.setRequestContext(ContractProto
                .RequestContext.newBuilder()
                .setNamespace("nsx")
                .build());
        Assert.assertEquals(false, mgr.getFilterChain("method1").doFilter(context1));

    }

    @Test
    public void testDeleteFilter() {
        mgr.AddFilter("method1", new NamespaceFilter());
        Context context = new Context("001");
        context.setRequestContext(ContractProto
                .RequestContext.newBuilder()
                .setNamespace("ns1")
                .build());

        mgr.removeFilter("method1", "NamespaceFilter");

        Assert.assertEquals(null, mgr.getFilter("method1", "NamespaceFilter"));
    }

    class NamespaceFilter implements Filter{

        private Set<String> permittedNamespaces;

        protected void addPermittedNamespace(String namespace) {
            permittedNamespaces.add(namespace);
        }

        public NamespaceFilter() {
           this.permittedNamespaces = new HashSet<>();
           this.addPermittedNamespace("ns1");
           this.addPermittedNamespace("ns2");
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
            return this.getClass().getSimpleName();
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
            return this.getClass().getSimpleName();
        }
    }
}
