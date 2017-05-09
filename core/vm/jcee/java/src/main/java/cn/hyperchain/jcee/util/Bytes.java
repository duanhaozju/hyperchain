package cn.hyperchain.jcee.util;

import com.esotericsoftware.kryo.Kryo;
import com.esotericsoftware.kryo.io.ByteBufferOutput;
import com.esotericsoftware.kryo.io.Input;
import com.esotericsoftware.kryo.io.Output;

/**
 * Created by wangxiaoyi on 2017/5/8.
 */
public class Bytes {

    public static byte[] toByteArray(String data){
        return data.getBytes();
    }

    public static byte[] toByteArray(char data) {
       return Character.toString(data).getBytes();
    }

    public static byte[] toByteArray(short data) {
        return Short.toString(data).getBytes();
    }

    public static byte[] toByteArray(int data) {
        return Integer.toString(data).getBytes();
    }

    public static byte[] toByteArray(long data) {
        return Long.toString(data).getBytes();
    }

    public static byte[] toByteArray(float data) {
        return Float.toString(data).getBytes();
    }

    public static byte[] toByteArray(double data) {
        return Double.toString(data).getBytes();
    }

    public static byte[] toByteArray(Object data) {
        Kryo kryo = new Kryo();
        Output output = new Output(new ByteBufferOutput());
        kryo.writeObject(output, data);
        return output.toBytes();
    }

    public static byte[] toByteArray(boolean data) {
        return Boolean.toString(data).getBytes();
    }

    public static int toInt(byte[] data) {
        return Integer.parseInt(new String(data));
    }

    public static long toLong(byte[] data) {
        return Long.parseLong(new String(data));
    }

    public static int toChar(byte[] data) {
        String tmp = new String(data);
        return tmp.charAt(0);
    }

    public static short toShort(byte[] data) {
        return Short.parseShort(new String(data));
    }

    public static float toFloat(byte[] data) {
        return Float.parseFloat(new String(data));
    }

    public static double toDouble(byte[] data) {
        return Double.parseDouble(new String(data));
    }

    public static String toString(byte[] data) {
        return new String(data);
    }

    public static <T> T toObject(byte[] data, Class<T> clazz) {
        Kryo kryo = new Kryo();
        Input input = new Input(data);
        return kryo.readObject(input, clazz);
    }
}
