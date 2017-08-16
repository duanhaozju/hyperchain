package cn.hyperchain.jcee.server.ledger;

import cn.hyperchain.jcee.client.ledger.Result;
import cn.hyperchain.jcee.common.exception.HyperjvmRuntimeException;
import com.google.protobuf.ByteString;
import org.junit.Assert;
import org.junit.Test;

/**
 * Created by wangxiaoyi on 2017/6/7.
 */
public class TestResult {

    @Test
    public void testIsEmpty() {
        Result result = new Result(ByteString.EMPTY);
        Assert.assertEquals(true, result.isEmpty());
        try {
            result.toBoolean();
        }catch (HyperjvmRuntimeException hre) {

        }
    }

    @Test
    public void testToFloat() {
        Result result = new Result(ByteString.copyFrom(new String("0.00234").getBytes()));
        Assert.assertEquals(0.00234f, result.toFloat(), 0.00001);
    }

    @Test
    public void testToStringAndBytes() {
        String x = "jsjsjjsjsj";
        Result result = new Result(ByteString.copyFrom(x.getBytes()));
        Assert.assertEquals(x, result.toString());

        Assert.assertArrayEquals(x.getBytes(), result.toBytes());
    }
}
