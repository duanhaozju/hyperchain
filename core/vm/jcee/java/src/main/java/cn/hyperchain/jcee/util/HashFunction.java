/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.util;

import org.apache.commons.codec.binary.Hex;
import scorex.crypto.hash.Keccak256;


/**
 * Created by wangxiaoyi on 2017/4/27.
 */
public class HashFunction {

    /**
     * compute the code hash using algo "KECCAK256"
     * @param code
     * @return
     */
    public static String computeCodeHash(byte[] code) {
        return Hex.encodeHexString(Keccak256.apply(code));
    }

}
