package cn.hyperchain.jcee.util;

import org.bouncycastle.util.encoders.Hex;
import org.junit.Test;

/**
 * Created by wangxiaoyi on 2017/4/27.
 */
public class TestIoHelper {

    @Test
    public void testReadCode() {
        String contractDir = TestIoHelper.class.getResource("/contracts").getPath();
        byte [] data = IOHelper.readCode(contractDir);
        System.out.println(Hex.toHexString(data));
        System.out.println(HashFunction.computeCodeHash(data));
    }

    @Test
    public void testCodeReadOrder() {
        String dir = TestIoHelper.class.getResource("/").getPath();
        byte [] data = IOHelper.readCode(dir);
        System.out.println(Hex.toHexString(data));
    }
}