/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.executor;

import cn.hyperchain.jcee.contract.ContractBase;
import cn.hyperchain.protos.ContractProto;

import java.util.LinkedList;
import java.util.List;

public class InvokeTask extends Task{

    public InvokeTask(ContractBase contract, ContractProto.Request request, Context context) {
        super(contract, request, context);
    }

    @Override
    public ContractProto.Response execute() {
        String funcName = request.getArgs(0).toStringUtf8();

        List<String> args = new LinkedList<>();
        for(int i = 1; i < request.getArgsList().size(); ++ i){
            args.add(request.getArgs(i).toStringUtf8());
        }
        ContractProto.Response r = ContractProto.Response.newBuilder()
                .setOk(contract.Invoke(funcName, args))
                .build();
        return r;
    }
}
