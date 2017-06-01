package cn.hyperchain.jcee.common.exception;

/**
 * Created by Think on 5/31/17.
 */
public class NotExistException extends HyperjvmException {
    public NotExistException() {
        super();
    }
    public NotExistException(String message) {
        super(message + " is not exist.");
    }
}
