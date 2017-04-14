/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.executor;

import cn.hyperchain.jcee.contract.ContractBase;
import cn.hyperchain.protos.Request;
import cn.hyperchain.protos.Response;

import java.util.LinkedList;
import java.util.List;

public class InvokeTask extends Task{

    public InvokeTask(ContractBase contract, Request request, Context context) {
        super(contract, request, context);
    }

    @Override
    public Response execute() {
        String funcName = request.getArgs(0).toStringUtf8();

        List<String> args = new LinkedList<>();
        for(int i = 1; i < request.getArgsList().size(); ++ i){
            args.add(request.getArgs(i).toStringUtf8());
        }
        Response r = Response.newBuilder()
                .setOk(contract.Invoke(funcName, args))
                .build();
        return r;
    }
}
