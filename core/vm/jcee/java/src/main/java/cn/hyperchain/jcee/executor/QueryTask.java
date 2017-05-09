/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.executor;

import cn.hyperchain.jcee.contract.ContractBase;
import cn.hyperchain.protos.ContractProto;
import com.google.protobuf.ByteString;

import java.util.LinkedList;
import java.util.List;

public class QueryTask extends Task{

    public QueryTask(ContractBase contract, ContractProto.Request request, Context context) {
        super(contract, request, context);
    }

    @Override
    public ContractProto.Response execute() {
        String funcName = request.getArgs(0).toStringUtf8();
        List<String> args = new LinkedList<>();
        for(int i = 1; i < request.getArgsList().size(); ++ i){
            args.add(request.getArgs(i).toStringUtf8());
        }

        ByteString bs = contract.Query(funcName, args);


        ContractProto.Response.Builder builder = ContractProto.Response.newBuilder();
        ContractProto.Response rs;
        if(bs == null) {
            builder.setOk(false);
        }else {
            builder.setOk(true);
            builder.setResult(bs);
        }
        //builder.setCid(request.getTxid());
        builder.setCodeHash(contract.getInfo().getCodeHash());
        rs = builder.build();
        return rs;
    }
}
