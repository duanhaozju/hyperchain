package cn.hyperchain.jcee.security;

import cn.hyperchain.jcee.contract.security.ByteCodeChecker;
import org.junit.Assert;
import org.junit.Test;

import java.io.*;

/**
 * Created by Think on 4/27/17.
 */
public class TestSecurity {

    private final String dir = "/src/test/resources/contracts/security";
    private final String contract1 = "/Contract1.class";
    private final String contract2 = "/Contract2.class";
    private final String contract3 = "/Contract3.class";
    private final String contract4 = "/Contract4.class";

    @Test
    public void testC1() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker byteCodeChecker = new ByteCodeChecker();
        Assert.assertTrue(byteCodeChecker.pass(getClass(contractDir + contract1)));
    }

    @Test
    public void testC2() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker byteCodeChecker = new ByteCodeChecker();
        Assert.assertFalse(byteCodeChecker.pass(getClass(contractDir + contract2)));
    }

    @Test
    public void testC3() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker byteCodeChecker = new ByteCodeChecker();
        Assert.assertFalse(byteCodeChecker.pass(getClass(contractDir + contract3)));
    }

    @Test
    public void testC4() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker byteCodeChecker = new ByteCodeChecker();
        Assert.assertFalse(byteCodeChecker.pass(getClass(contractDir + contract4)));
    }

    public byte[] getClass(String path) {
        ByteArrayOutputStream out = new ByteArrayOutputStream(1024);
        try {
            FileInputStream in = new FileInputStream(path);
            byte[] temp = new byte[1024];
            int size = 0;
            while((size = in.read(temp))!=-1) {
                out.write(temp, 0, size);
            }
            in.close();
        } catch (FileNotFoundException e) {
            e.printStackTrace();
        } catch (IOException e) {
            e.printStackTrace();
        }
        return out.toByteArray();
    }
}


