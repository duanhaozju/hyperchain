/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract.security;

/**
 * Created by wangxiaoyi on 2017/4/25.
 */
public interface Checker {
    boolean pass(byte[] clazz);
    boolean pass(String absoluteClassPath);
    boolean passAll(String absoluteDirPath);
}
