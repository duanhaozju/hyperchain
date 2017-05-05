package cn.hyperchain.jcee.security;

import cn.hyperchain.jcee.contract.security.ByteCodeChecker;
import org.junit.Assert;
import org.junit.Test;

import java.io.*;

/**
 * Created by Think on 4/27/17.
 */
public class TestSecurity {

    private String dir = "/src/test/resources/contracts/security";
    private String normalContract = "/NormalContract.class";
    private String ioContract = "/IOContract.class";
    private String threadContract = "/ThreadContract.class";
    private String syncContract = "/SyncContract.class";
    private String reflectContract = "/ReflectContract.class";
    private String concurrentContract = "/ConcurrentContract.class";
    private String netContract = "/NetContract.class";
    private String syncContract2 = "/SyncContract2.class";

    private String userPath = "/securityRule.yaml";

    @Test
    public void testNormalContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker byteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertTrue(byteCodeChecker.pass(getClass(contractDir + normalContract)));
        ByteCodeChecker byteCodeChecker1 = new ByteCodeChecker();
        Assert.assertTrue(byteCodeChecker1.pass(getClass(contractDir + normalContract)));
    }

    @Test
    public void testIOContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker byteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertFalse(byteCodeChecker.pass(getClass(contractDir + ioContract)));
    }

    @Test
    public void testThreadContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker byteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertFalse(byteCodeChecker.pass(getClass(contractDir + threadContract)));
    }

    @Test
    public void testSyncContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker byteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertFalse(byteCodeChecker.pass(getClass(contractDir + syncContract)));
    }

    @Test
    public void testReflectContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker byteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        byteCodeChecker.pass(contractDir + reflectContract);
        Assert.assertFalse(byteCodeChecker.pass(contractDir + reflectContract));
    }

    @Test
    public void testConcurrentContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker byteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        byteCodeChecker.pass(contractDir + concurrentContract);
        Assert.assertFalse(byteCodeChecker.pass(contractDir + concurrentContract));
    }

    @Test
    public void testNetContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker byteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        byteCodeChecker.pass(contractDir + netContract);
        Assert.assertFalse(byteCodeChecker.pass(contractDir + netContract));
    }

    @Test
    public void testSyncContract2() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker byteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        byteCodeChecker.pass(contractDir + syncContract2);
        Assert.assertFalse(byteCodeChecker.pass(contractDir + syncContract2));
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


