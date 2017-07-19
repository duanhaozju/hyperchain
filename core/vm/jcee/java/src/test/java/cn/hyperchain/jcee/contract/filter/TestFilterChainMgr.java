package cn.hyperchain.jcee.contract.filter;

import cn.hyperchain.jcee.executor.Context;
import org.junit.AfterClass;
import org.junit.BeforeClass;
import org.junit.Test;

/**
 * Created by wangxiaoyi on 2017/7/18.
 */
public class TestFilterChainMgr {

    FilterChainMgr mgr;

    @BeforeClass
    public void setup() {
        mgr = new FilterChainMgr();
    }

    @AfterClass
    public void destroy() {

    }

    @Test
    public void testAddGetFilterChain() {
        FilterChain fc = new FilterChain();
        fc.addFilter(new TestFilter());

    }

    class TestFilter implements Filter{
        @Override
        public boolean doFilter(Context context) {
            return false;
        }

        @Override
        public String getName() {
            return "TestFilter";
        }
    }
}
