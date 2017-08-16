package cn.hyperchain.jcee.common;

/**
 * Created by wangxiaoyi on 2017/6/27.
 */
public interface Coder {
    String encode(String data);
    String encode(byte[] data);

    String decode(String codeData);
    String decode(byte[] codeData);
}
