package cn.hyperchain.jcee.client.executor;

public abstract class AbstractContractHandler {

    protected static volatile AbstractContractHandler ch;

    public static final AbstractContractHandler getContractHandler() {
        if (ch == null) {
            throw new NullPointerException("ContractHandler is not initialized after bootstrap");
        }
        return ch;
    }

    public abstract void addHandler(String namespace);

    public abstract boolean hasHandlerForNamespace(String namespace);

    public abstract IHandler get(String namespace);
}

