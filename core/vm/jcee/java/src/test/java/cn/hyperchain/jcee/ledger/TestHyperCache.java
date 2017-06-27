package cn.hyperchain.jcee.ledger;

import com.google.protobuf.ByteString;
import org.junit.Assert;
import org.junit.Test;

/**
 * Created by huhu on 2017/6/12.
 */
public class TestHyperCache {
    @Test
    public void testPutAndGet(){
        HyperCache hc = new HyperCache();

        String key = "testkey";
        int value =10;

        byte[] realK = key.getBytes();

        hc.put(realK,new Integer(value).toString().getBytes());

        Result result = new Result(ByteString.copyFrom(hc.get(realK)));
        Assert.assertEquals(result.toInt(),value);
    }

    @Test
    public void testMixKey(){
        HyperCache hc = new HyperCache();

        MixThread m1 = new MixThread("global_cid1_",hc,"value1");
        MixThread m2 = new MixThread("global_cid2_",hc,"value2");
        new Thread(m1).start();
        new Thread(m2).start();
        try {
            Thread.sleep(10);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }

    }

    class MixThread implements Runnable{
        private String prefix;
        private HyperCache hc;
        private String value;
        public MixThread(String prefix,HyperCache hc, String value){
            this.prefix = prefix;
            this.hc = hc;
            this.value = value;
        }
        public void run(){
            while (true){
                String key = prefix+"key1";
                byte[] realK = key.getBytes();
                hc.put(realK, value.getBytes());

                Result result = new Result(ByteString.copyFrom(hc.get(realK)));
                Assert.assertEquals(result.toString(),value);
            }
        }
    }

    @Test
    public void TestBytes(){
        String str = "hello";
        System.out.println(str.getBytes().equals(str.getBytes()));
        byte[] key1 = str.getBytes();
        byte[] key2 = str.getBytes();
        ByteString bs1 = ByteString.copyFrom(key1);
        ByteString bs2 = ByteString.copyFrom(key2);
        System.out.println(bs1.toStringUtf8());
        System.out.println(bs2.toStringUtf8());
        System.out.println(bs2.equals(bs1));

        byte[] b1 = bs1.toByteArray();
        byte[] b2 = bs2.toByteArray();
        System.out.println(b1.equals(b2));
    }

}
