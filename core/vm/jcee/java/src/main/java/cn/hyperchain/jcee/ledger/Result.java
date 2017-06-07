package cn.hyperchain.jcee.ledger;

import cn.hyperchain.jcee.util.Bytes;
import com.google.protobuf.ByteString;

/**
 * Created by huhu on 2017/6/7.
 */
public class Result {
    private ByteString value;

    public Result(ByteString value){
        this.value = value;
    }

    public boolean toBoolean() {
        return Boolean.parseBoolean(value.toStringUtf8());
    }

    public char toChar() {
        return value.toStringUtf8().charAt(0);
    }

    public short toShort() {
        return Short.parseShort(value.toStringUtf8());
    }

    public int toInt() {
        return Integer.parseInt(value.toStringUtf8());
    }

    public float toFloat() {
        return Float.parseFloat(value.toStringUtf8());
    }

    public double toDouble() {
        return Double.parseDouble(value.toStringUtf8());
    }

    public String transferToString() {
        return value.toStringUtf8();
    }

    public <T> T toObeject(Class<T> clazz){
        return Bytes.toObject(value.toByteArray(), clazz);
    }

    public boolean isEmpty(){
        return value.isEmpty();
    }

    public ByteString getValue() {
        return value;
    }
}