package cn.hyperchain.jcee.server.contract;

import cn.hyperchain.jcee.common.Event;
import org.junit.Test;

/**
 * Created by wangxiaoyi on 2017/6/26.
 */
public class TestEvent {

    @Test
    public void TestToJsonString() {
        Event event = new Event("contract");
        event.put("method", "invoke");
        System.out.println(event);

        Event event2 = new Event("contract_event", "test", "deploy");
        event2.put("time used", "112");
        event2.put("any debug", "111");
        System.out.println(event2);
    }


}
