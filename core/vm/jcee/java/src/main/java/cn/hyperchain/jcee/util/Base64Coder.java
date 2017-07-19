package cn.hyperchain.jcee.util;

import java.util.Base64;

/**
 * Created by wangxiaoyi on 2017/6/27.
 */
public class Base64Coder implements Coder{

    @Override
    public String encode(String data) {
       return encode(data.getBytes());
    }

    @Override
    public String encode(byte[] data) {
        return new String(Base64.getEncoder().encode(data));
    }

    @Override
    public String decode(String codeData) {
        return new String(Base64.getDecoder().decode(codeData));
    }

    @Override
    public String decode(byte[] codeData) {
        return new String(Base64.getDecoder().decode(codeData));
    }
}
