/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.executor;

import cn.hyperchain.jcee.contract.ContractBase;
import cn.hyperchain.protos.Request;
import cn.hyperchain.protos.Response;
import com.google.protobuf.ByteString;

import java.util.LinkedList;
import java.util.List;

public class QueryTask extends Task{

    public QueryTask(ContractBase contract, Request request, Context context) {
        super(contract, request, context);
    }

    @Override
    public Response execute() {
        String funcName = request.getArgs(0).toStringUtf8();
        List<String> args = new LinkedList<>();
        for(int i = 1; i < request.getArgsList().size(); ++ i){
            args.add(request.getArgs(i).toStringUtf8());
        }

        ByteString bs = contract.Query(funcName, args);

        Response.Builder builder = Response.newBuilder();
        Response rs;
        if(bs == null) {
            builder.setOk(false);
        }else {
            builder.setOk(true);
            builder.setResult(bs);
        }
        builder.setId(request.getTxid());
        rs = builder.build();
        return rs;
    }
}
