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
    private String reflectContract2 = "/ReflectContract2.class";
    private String concurrentContract = "/ConcurrentContract.class";
    private String netContract = "/NetContract.class";
    private String syncContract2 = "/SyncContract2.class";
    private String normalData = "/NormalData.class";
    private String timeContract = "/TimeContract.class";
    private String timeContract2 = "/TimeContract2.class";

    private String userPath = "/securityRule.yaml";

    @Test
    public void testNormalContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker userByteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertTrue(userByteCodeChecker.pass(getClass(contractDir + normalContract)));
        ByteCodeChecker systemByteCodeChecker = new ByteCodeChecker();
        Assert.assertTrue(systemByteCodeChecker.pass(getClass(contractDir + normalContract)));
        Assert.assertFalse(systemByteCodeChecker.passAll(contractDir));
    }

    @Test
    public void testIOContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker systemByteCodeChecker = new ByteCodeChecker();
        ByteCodeChecker userByteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertFalse(systemByteCodeChecker.pass(getClass(contractDir + ioContract)));
        Assert.assertTrue(userByteCodeChecker.pass(getClass(contractDir + ioContract)));
    }

    @Test
    public void testThreadContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker systemByteCodeChecker = new ByteCodeChecker();
        ByteCodeChecker userByteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertFalse(systemByteCodeChecker.pass(getClass(contractDir + threadContract)));
        Assert.assertTrue(userByteCodeChecker.pass(getClass(contractDir + threadContract)));
    }

    @Test
    public void testSyncContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker systemByteCodeChecker = new ByteCodeChecker();
        ByteCodeChecker userByteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertFalse(systemByteCodeChecker.pass(getClass(contractDir + syncContract)));
        Assert.assertTrue(userByteCodeChecker.pass(getClass(contractDir + syncContract)));
    }

    @Test
    public void testReflectContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker systemByteCodeChecker = new ByteCodeChecker();
        ByteCodeChecker userByteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertFalse(systemByteCodeChecker.pass(contractDir + reflectContract));
        Assert.assertTrue(userByteCodeChecker.pass(contractDir + reflectContract));
    }

    @Test
    public void testReflectContract2() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker systemByteCodeChecker = new ByteCodeChecker();
        ByteCodeChecker userByteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertFalse(systemByteCodeChecker.pass(contractDir + reflectContract2));
        Assert.assertTrue(userByteCodeChecker.pass(contractDir + reflectContract2));
    }

    @Test
    public void testConcurrentContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker systemByteCodeChecker = new ByteCodeChecker();
        ByteCodeChecker userByteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertFalse(systemByteCodeChecker.pass(contractDir + concurrentContract));
        Assert.assertTrue(userByteCodeChecker.pass(contractDir + concurrentContract));
    }

    @Test
    public void testNetContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker systemByteCodeChecker = new ByteCodeChecker();
        ByteCodeChecker userByteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertFalse(systemByteCodeChecker.pass(contractDir + netContract));
        Assert.assertTrue(userByteCodeChecker.pass(contractDir + netContract));
    }

    @Test
    public void testSyncContract2() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker systemByteCodeChecker = new ByteCodeChecker();
        ByteCodeChecker userByteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertFalse(systemByteCodeChecker.pass(contractDir + syncContract2));
        Assert.assertTrue(userByteCodeChecker.pass(contractDir + syncContract2));
    }

    @Test
    public void testNormalData() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker systemByteCodeChecker = new ByteCodeChecker();
        ByteCodeChecker userByteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertTrue(systemByteCodeChecker.pass(contractDir + normalData));
        Assert.assertTrue(userByteCodeChecker.pass(contractDir + normalData));
    }

    @Test
    public void testTimeContract() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker systemByteCodeChecker = new ByteCodeChecker();
        ByteCodeChecker userByteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertFalse(systemByteCodeChecker.pass(contractDir + timeContract));
        Assert.assertTrue(userByteCodeChecker.pass(contractDir + timeContract));
    }

    @Test
    public void testTimeContract2() {
        String contractDir = System.getProperty("user.dir") + dir;
        ByteCodeChecker systemByteCodeChecker = new ByteCodeChecker();
        ByteCodeChecker userByteCodeChecker = new ByteCodeChecker(contractDir + userPath);
        Assert.assertFalse(systemByteCodeChecker.pass(contractDir + timeContract2));
        Assert.assertTrue(userByteCodeChecker.pass(contractDir + timeContract2));
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


