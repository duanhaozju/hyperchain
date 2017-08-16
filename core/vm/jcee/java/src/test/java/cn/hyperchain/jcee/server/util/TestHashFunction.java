package cn.hyperchain.jcee.server.util;

import com.google.protobuf.ByteString;
import org.junit.Test;

import java.nio.charset.Charset;

/**
 * Created by wangxiaoyi on 2017/4/27.
 */
public class TestHashFunction {

    @Test
    public void test() throws Exception{
        ByteString bs = ByteString.copyFrom(new String("1234"), Charset.defaultCharset());
//        double x = DoubleValue.parseFrom(bs).getValue();
//        System.out.println(x);

        double x = Double.parseDouble(bs.toStringUtf8());
        System.out.println(x);


        boolean b = true;
        System.out.println(true == Boolean.parseBoolean(new String(Boolean.toString(b).getBytes())));

    }
}
