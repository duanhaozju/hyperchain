/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.executor;


import cn.hyperchain.protos.ContractProto;

//Context used to contain info shared among a session.
public class Context {
    private String id;
    private ContractProto.RequestContext requestContext;

    public Context(String id) {
        this.id = id;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public ContractProto.RequestContext getRequestContext() {
        return requestContext;
    }

    public void setRequestContext(ContractProto.RequestContext requestContext) {
        this.requestContext = requestContext;
    }
}
