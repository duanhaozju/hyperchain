package cn.hyperchain.jcee.server.common.exception;

/**
 * Created by Think on 5/31/17.
 */
public class ResultNotExistException extends HyperjvmRuntimeException {
    public ResultNotExistException() {
        super();
    }
    public ResultNotExistException(String result) {
        super("Error " + result + " is not exist.");
    }
}
