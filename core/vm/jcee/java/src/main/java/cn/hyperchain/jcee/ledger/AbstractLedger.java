package cn.hyperchain.jcee.ledger;

import cn.hyperchain.jcee.executor.Context;

public abstract class AbstractLedger implements ILedger{

    private ThreadLocal<Context> contexts;

    public AbstractLedger(){
        contexts = new ThreadLocal<>();
    }

    public void setContext(Context context) {
        contexts.set(context);
    }

    public void removeContext() {
        contexts.remove();
    }

    public Context getContext() {
        return contexts.get();
    }
}
