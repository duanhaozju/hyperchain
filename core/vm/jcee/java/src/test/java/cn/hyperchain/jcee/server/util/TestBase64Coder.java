package cn.hyperchain.jcee.server.util;

import cn.hyperchain.jcee.common.Base64Coder;
import cn.hyperchain.jcee.common.Coder;
import cn.hyperchain.jcee.common.Event;
import org.junit.Assert;
import org.junit.Test;

/**
 * Created by wangxiaoyi on 2017/6/27.
 */
public class TestBase64Coder {

    @Test
    public void testEncodeAndDecode() {
        Coder coder = new Base64Coder();
        Event event = new Event("event1");
        event.addTopic("t1");
        event.addTopic("t2");

        event.put("attr1", "v1");
        event.put("attr2", "v2");

        Assert.assertEquals(event.toString(), coder.decode(coder.encode(event.toString())));

    }
}
