package cn.hyperchain.jcee.executor;

public abstract class IContractHandler {

    protected static volatile IContractHandler ch;

    public static IContractHandler getContractHandler() {
        if (ch == null) {
            throw new NullPointerException("ContractHandler is not initialized after bootstrap");
        }
        return ch;
    }

    public abstract void addHandler(String namespace);

    public abstract boolean hasHandlerForNamespace(String namespace);

    public abstract IHandler get(String namespace);
}

