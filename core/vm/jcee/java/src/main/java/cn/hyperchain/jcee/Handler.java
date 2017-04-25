/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee;

import cn.hyperchain.jcee.contract.ContractInfo;
import cn.hyperchain.jcee.contract.ContractManager;
import cn.hyperchain.jcee.executor.*;
import cn.hyperchain.protos.Request;
import cn.hyperchain.protos.RequestContext;
import cn.hyperchain.protos.Response;
import com.google.protobuf.ByteString;
import io.grpc.stub.StreamObserver;
import org.apache.log4j.Logger;

import java.io.FileInputStream;
import java.io.IOException;
import java.util.List;
import java.util.Properties;
import java.util.concurrent.Future;

/**
 * Handler used to handle the real request
 */
public class Handler {

    private Logger logger = Logger.getLogger(Handler.class.getSimpleName());
    private ContractManager cm;
    private ContractExecutor executor;

    public Handler(){
        cm = new ContractManager();
        executor = new ContractExecutor();
    }

    /**
     * handle query method
     * @param request query request
     * @param responseObserver
     */
    public void query(Request request, StreamObserver<Response> responseObserver){
        Task task = new QueryTask(cm.getContract(request.getContext().getCid()), request, constructContext(request.getContext()));
        Future<Response> future = executor.execute(task);
        Response response = null;
        try{
            response = future.get();
            logger.info(response.getResult().toStringUtf8());
        }catch (Exception e) {
            logger.error(e);
            response = Response.newBuilder().setOk(false)
                    //.setId(request.getTxid())
                    .setResult(ByteString.copyFromUtf8(e.getMessage()))
                    .build();
        }finally {
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
    }

    /**
     * invoke the contract method
     * @param request invoke method
     * @param responseObserver
     */
    public void invoke(Request request, StreamObserver<Response> responseObserver){
        logger.debug("cid is " + request.getContext().getCid());
        logger.debug("contract is " + cm.getContract(request.getContext().getCid()));
        Task task = new InvokeTask(cm.getContract(request.getContext().getCid()), request, constructContext(request.getContext()));
        Future<Response> future = executor.execute(task);
        Response response = null;
        try{
            response = future.get();
        }catch (Exception e) {
            logger.error(e);
            e.printStackTrace();
            response = Response.newBuilder().setOk(false)
                    //.setId(request.getTxid())
                    .setResult(ByteString.copyFromUtf8(e.getMessage()))
                    .build();
        }finally {
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
    }

    /**
     * deploy the contract
     * @param request deploy request
     * @param responseObserver
     */
    public void deploy(Request request, StreamObserver<Response> responseObserver){

        List<ByteString>  args = request.getArgsList();
        if (args == null || args.size() == 0) {
            logger.error("deploy failed, invalid num of deploy args");
        }
        String contractPath = args.get(0).toStringUtf8();
        Properties props = new Properties();
        try {
            props.load(new FileInputStream(contractPath + "/contract.properties"));
        } catch (IOException e) {
            e.printStackTrace();
        }
        ContractInfo info = null;
        Response r = null;
        if (props != null) {
            info = new ContractInfo(props.getProperty(Constants.CONTRACT_NAME), request.getContext().getCid(), "0xx");
            info.setContractPath(contractPath);
            info.setClassPrefix(props.getProperty(Constants.CONTRACT_CLASS_SUPER_DIR));
            info.setContractMainName(props.getProperty(Constants.CONTRACT_MAIN_CLASS));
            info.setId(request.getContext().getCid());
            info.setNamespace(request.getContext().getNamespace());
            info.setCreateTime(System.currentTimeMillis());
            info.setModifyTime(info.getCreateTime());
            boolean extractSuccess = extractConstructorArgs(info, request.getArgsList());
            if (extractSuccess) {
                logger.debug(info);
                boolean rs = cm.deployContract(info);
                if (rs == true) {
                    r = Response.newBuilder().setOk(rs).build();
                    //add code hash
                } else {
                    r = Response.newBuilder().setOk(rs).build();
                }
            }else {
                r = Response.newBuilder().setOk(false).build();
            }
        }else  {
            r = Response.newBuilder().setOk(false).build();
        }
        if (responseObserver != null) {
            responseObserver.onNext(r);
            responseObserver.onCompleted();
        }
    }

    public ContractManager getContractMgr() {
        return cm;
    }

    public Context constructContext(RequestContext context) {
        Context ct = new Context(context.getTxid());
        ct.setRequestContext(context);
        return ct;
    }

    /**
     * extractConstructorArgs
     * @param info contract info.
     * @param args contract constructor type and the related value.
     * @return extract status.
     */
    public boolean extractConstructorArgs(ContractInfo info, List<ByteString> args) {
        int n = args.size();
        int i = 1, j = (n - i) / 2 + 1; // TODO: may not start from 1
        Class argClasses[] = new Class[j - 1];
        Object objectArgs[] = new Object[j - 1];
        while (j < n) {
            String className = args.get(i).toStringUtf8();
            String arg = args.get(j).toStringUtf8();
            switch (className){
                case "boolean":
                    argClasses[i - 1] = boolean.class;
                    objectArgs[i - 1] = arg == "true" ? true : false;
                    break;
                case "char":
                     argClasses[i - 1] = char.class;
                     objectArgs[i - 1] = arg.charAt(0);
                     break;
                case "short":
                    argClasses[i - 1] = short.class;
                    objectArgs[i - 1] = Short.parseShort(arg);
                    break;
                case "int":
                    argClasses[i - 1] = int.class;
                    objectArgs[i - 1] = Integer.parseInt(arg);
                    break;
                case "long":
                    argClasses[i - 1] = long.class;
//                    System.out.print(arg);
                    objectArgs[i - 1] = Long.parseLong(arg);
                    break;
                case "float":
                    argClasses[i - 1] = float.class;
                    objectArgs[i - 1] = Float.parseFloat(arg);
                    break;
                case "double":
                    argClasses[i - 1] = double.class;
                    objectArgs[i - 1] = Double.parseDouble(arg);
                    break;
                case "String":
                    argClasses[i - 1] = String.class;
                    objectArgs[i - 1] = arg;
                    break;
                default:
                   logger.error("can not pass non string object to contract constructor");
                   return false;
            }
            i ++;
            j ++;
        }
        info.setArgClasses(argClasses);
        info.setArgs(objectArgs);
        return true;
    }
}
