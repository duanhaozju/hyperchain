/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.executor;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import cn.hyperchain.protos.ContractProto;

import java.util.LinkedList;
import java.util.List;

public class InvokeTask extends Task{

    public InvokeTask(ContractTemplate contract, ContractProto.Request request, Context context) {
        super(contract, request, context);
    }

    @Override
    public ContractProto.Response execute() {
        System.out.println(request.getArgsCount());
        String funcName = request.getArgs(0).toStringUtf8();

        List<String> args = new LinkedList<>();
        for(int i = 0; i < request.getArgsList().size(); ++ i){
            if (i > 0) args.add(request.getArgs(i).toStringUtf8());
        }
        ExecuteResult result = contract.invoke(funcName, args);

        ContractProto.Response r = ContractProto.Response.newBuilder()
                .setOk(result.isSuccess())
                .setResult(result.getResultByteString())
                .setCodeHash(contract.getInfo().getCodeHash())
                .build();
        return r;
    }
}
