package cn.hyperchain.jcee.util;

import junitparams.JUnitParamsRunner;
import junitparams.Parameters;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;

/**
 * Created by wangxiaoyi on 2017/5/8.
 */
@RunWith(JUnitParamsRunner.class)
public class TestBytes {

    @Test
    @Parameters({
            "1",
            "1222222222",
            "-1",
    })
    public void intBytesOp(int obj) {
        byte[] data = Bytes.toByteArray(obj);
        assert Integer.parseInt(Bytes.toString(data)) == obj;
    }

    @Test
    public void testObjectSerialize() {
        Person person = new Person(123, "wxu");
        byte[] data = Bytes.toByteArray(person);
        Person deserPerson = Bytes.toObject(data, Person.class);
        Assert.assertEquals(true, person.equals(deserPerson));
    }

    @Test
    public void testObjectSerializePerformance() {
        long t1 = System.currentTimeMillis();
        for (int i = 0; i < 10 * 10000; i ++) {
            Person person = new Person(123, "wxu");
            byte[] data = Bytes.toByteArray(person);
            Person deserPerson = (Person) Bytes.toObject(data, Person.class);
        }
        long t2 = System.currentTimeMillis();
        System.out.println(100000 * 1.0 / (t2 - t1) * 1000);
    }

    @Data
    @AllArgsConstructor
    @NoArgsConstructor
    static class Person{
        int age;
        String name;
    }
}
