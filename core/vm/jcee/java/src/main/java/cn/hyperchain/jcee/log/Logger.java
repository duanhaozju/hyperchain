/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.log;

/**
 * Created by Think on 6/6/17.
 */
public interface Logger {
    void debug(Object message);
    void warn(Object message);
    void info(Object message);
    void error(Object message);
    void fatal(Object message);
}
