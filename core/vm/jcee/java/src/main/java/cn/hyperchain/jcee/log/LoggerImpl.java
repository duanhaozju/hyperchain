/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.log;

/**
 * Created by Think on 6/6/17.
 */
public class LoggerImpl implements Logger {
    private org.apache.log4j.Logger logger;

    public LoggerImpl(String name) {
        logger = org.apache.log4j.Logger.getLogger(name);
    }

    @Override
    public void debug(Object message) {
        logger.debug(message);
    }

    @Override
    public void warn(Object message) {
        logger.warn(message);
    }

    @Override
    public void info(Object message) {
        logger.info(message);
    }

    @Override
    public void error(Object message) {
        logger.error(message);
    }

    @Override
    public void fatal(Object message) {
        logger.fatal(message);
    }

}
