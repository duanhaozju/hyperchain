package cn.hyperchain.jcee.common;

import cn.hyperchain.jcee.util.Bytes;
import com.google.protobuf.ByteString;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * Created by wangxiaoyi on 2017/5/9.
 */
@Data
@NoArgsConstructor
@AllArgsConstructor
//ExecuteResult used to encapsulate the consequence of some calculation.
public class ExecuteResult<T> {

    private boolean success = false;

    //this result is used to store execute result if success = true
    //                       store error message if success = false
    private T result;

    public ByteString getResultByteString() {
        if (result == null) {
            return null;
        }
        if (result instanceof Character) {
            return ByteString.copyFrom(Bytes.toByteArray(((Character) result).charValue()));
        }else if (result instanceof Short) {
            return ByteString.copyFrom(Bytes.toByteArray(((Short) result).shortValue()));
        }else if (result instanceof Integer) {
            return ByteString.copyFrom(Bytes.toByteArray(((Integer) result).intValue()));
        }else if (result instanceof Long) {
            return ByteString.copyFrom(Bytes.toByteArray(((Long) result).longValue()));
        }else if (result instanceof Float) {
            return ByteString.copyFrom(Bytes.toByteArray(((Float) result).floatValue()));
        }else if (result instanceof Double) {
            return ByteString.copyFrom(Bytes.toByteArray(((Double) result).doubleValue()));
        }else if (result instanceof Boolean) {
            return ByteString.copyFrom(Bytes.toByteArray(((Boolean) result).booleanValue()));
        }else if (result instanceof String) {
            return ByteString.copyFrom(Bytes.toByteArray((String) result));
        }else {
            return ByteString.copyFrom(Bytes.toByteArray(result));
        }
    }
}
