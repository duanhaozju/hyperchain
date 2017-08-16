package cn.hyperchain.jcee.server.util;

import cn.hyperchain.jcee.common.HashFunction;
import cn.hyperchain.jcee.server.common.IOHelper;
import org.apache.commons.codec.binary.Hex;
import org.junit.Test;

/**
 * Created by wangxiaoyi on 2017/4/27.
 */
public class TestIoHelper {

    @Test
    public void testReadCode() {
        String contractDir = TestIoHelper.class.getResource("/contracts").getPath();
        byte [] data = IOHelper.readCode(contractDir);
        System.out.println(Hex.encodeHexString(data));
        System.out.println(HashFunction.computeCodeHash(data));
    }

    @Test
    public void testCodeReadOrder() {
        String dir = TestIoHelper.class.getResource("/").getPath();
        byte [] data = IOHelper.readCode(dir);
        System.out.println(Hex.encodeHexString(data));
    }
}