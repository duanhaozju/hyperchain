package cn.hyperchain.jcee.ledger;

import cn.hyperchain.jcee.common.exception.ResultNotExistException;
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
        if (this.isEmpty()) {
            throw new ResultNotExistException();
        }
        return Boolean.parseBoolean(value.toStringUtf8());
    }

    public char toChar() {
        if (this.isEmpty()) {
            throw new ResultNotExistException();
        }
        return value.toStringUtf8().charAt(0);
    }

    public short toShort() {
        if (this.isEmpty()) {
            throw new ResultNotExistException();
        }
        return Short.parseShort(value.toStringUtf8());
    }

    public int toInt()  {
        if (this.isEmpty()) {
            throw new ResultNotExistException();
        }
        return Integer.parseInt(value.toStringUtf8());
    }

    public float toFloat() {
        if (this.isEmpty()) {
            throw new ResultNotExistException();
        }
        return Float.parseFloat(value.toStringUtf8());
    }

    public double toDouble() {
        if (this.isEmpty()) {
            throw new ResultNotExistException();
        }
        return Double.parseDouble(value.toStringUtf8());
    }

    public String toString() {
        if (this.isEmpty()) {
            throw new ResultNotExistException();
        }
        return value.toStringUtf8();
    }

    public <T> T toObeject(Class<T> clazz) {
        if (this.isEmpty()) {
            throw new ResultNotExistException();
        }
        return Bytes.toObject(value.toByteArray(), clazz);
    }

    public boolean isEmpty(){
        return value.isEmpty();
    }

    public ByteString getValue() {
        if (this.isEmpty()) {
            throw new ResultNotExistException();
        }
        return value;
    }

    public byte[] toBytes() {
        if (this.isEmpty()) {
            throw new ResultNotExistException();
        }
        return value.toByteArray();
    }
}