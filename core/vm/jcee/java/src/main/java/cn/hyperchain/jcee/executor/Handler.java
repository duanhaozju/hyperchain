/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.executor;

import cn.hyperchain.jcee.common.Constants;
import cn.hyperchain.jcee.contract.ContractState;
import cn.hyperchain.jcee.contract.ContractTemplate;
import cn.hyperchain.jcee.contract.ContractInfo;
import cn.hyperchain.jcee.contract.ContractManager;
import cn.hyperchain.jcee.util.Errors;
import cn.hyperchain.jcee.util.HashFunction;
import cn.hyperchain.jcee.util.IOHelper;
import cn.hyperchain.protos.ContractProto;
import com.google.protobuf.ByteString;
import io.grpc.stub.StreamObserver;
import org.apache.log4j.Logger;

import java.io.FileInputStream;
import java.io.IOException;
import java.util.List;
import java.util.Properties;

/**
 * Handler used to handle the real request
 * a namespace has a handler to handle request within this namespace
 */
public class Handler {

    private Logger logger = Logger.getLogger(Handler.class.getSimpleName());
    private ContractManager cm;

    enum TaskType {QUERY, INVOKE}

    public Handler(int ledgerPort){
        cm = new ContractManager(ledgerPort);
    }

    public void freeze(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver) {
        ContractProto.Response response;
        String cid = request.getContext().getCid();
        logger.info("System try to freeze contract with id " + cid);
        ContractInfo info = cm.getContractInfoByCid(cid);
        String err;
        if (info == null) {
            err = "freeze contract failed, contract with id " + cid + " is nonexistent";
            logger.error(err);
            Errors.ReturnErrMsg(err, responseObserver);
            return;
        }
        if (info.getState() != ContractState.FREEZE) {
            info.setState(ContractState.FREEZE);
        }
        response = ContractProto.Response.newBuilder()
                .setOk(true)
                .setCodeHash(info.getCodeHash())
                .build();
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    public void unfreeze(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver) {
        ContractProto.Response response;
        String cid = request.getContext().getCid();
        logger.info("System try to unfreeze contract with id " + cid);
        String err;
        ContractInfo info = cm.getContractInfoByCid(cid);
        if(info == null) {
            err = "unfreeze contract failed, contract with id " + cid + " is nonexistent";
            logger.error(err);
            Errors.ReturnErrMsg(err, responseObserver);
            return;
        }
        info.setState(ContractState.NORMAL);
        response = ContractProto.Response.newBuilder()
                .setOk(true)
                .setCodeHash(info.getCodeHash())
                .build();
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }


    public void destroy(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver) {
        /**
         * 1.remove the contract object from ContractManager
         */
        if(checkContractExistence(request, responseObserver)) {
              String cid = request.getContext().getCid();
              cm.destroyContract(cid);
        }
    }

    /**
     * update a contract is much more complex
     * 1.destroy old contract
     * 2.load new contract class and register which into the handler and ContractManager
     * @warn we should keep the update function args the same as deploy function args except for
     * @param request
     * @param responseObserver
     */
    public void update(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver) {

        if (checkContractExistence(request, responseObserver)) {
            //1.destroy old contract
            String cid = request.getContext().getCid();
            cm.destroyContract(cid);

            //2.deploy new contract
            destroy(request, responseObserver);
        }
    }

    /**
     * invoke the contract method
     * @param request invoke method
     * @param responseObserver
     */
    public void invoke(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver){
        ContractProto.Response response = null;
        logger.debug("cid is " + request.getContext().getCid());
        logger.debug("contract is " + cm.getContract(request.getContext().getCid()));
        Task task = constructTask(TaskType.INVOKE, request);
        if(task == null) {
            Errors.ReturnErrMsg("contract with id " + request.getContext().getCid() + " is not found", responseObserver);
            return;
        }
        try{
            response = task.call();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }catch (Exception e) {
            logger.error(e);
            Errors.ReturnErrMsg(e.getMessage(), responseObserver);
            return;
        }
    }

    /**
     * deploy the contract
     * @param request deploy request
     * @param responseObserver
     */
    public void deploy(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver){

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
        ContractProto.Response r;
        if (props != null) {
            info = new ContractInfo(props.getProperty(Constants.CONTRACT_NAME), request.getContext().getCid(), "0xx");
            info.setContractPath(contractPath);
            info.setClassPrefix(props.getProperty(Constants.CONTRACT_CLASS_SUPER_DIR));
            info.setContractMainName(props.getProperty(Constants.CONTRACT_MAIN_CLASS));
            info.setCid(request.getContext().getCid());
            info.setNamespace(request.getContext().getNamespace());
            info.setCreateTime(System.currentTimeMillis());
            info.setModifyTime(info.getCreateTime());
            caculateCodeHash(info);
            boolean extractSuccess = extractConstructorArgs(info, request.getArgsList());
            if (extractSuccess) {
                logger.debug(info);
                boolean rs = cm.deployContract(info);
                if (rs == true) {
                    r = ContractProto.Response.newBuilder()
                            .setOk(rs)
                            .setCodeHash(info.getCodeHash())
                            .build();
                } else {
                    r = ContractProto.Response.newBuilder().setCodeHash(info.getCodeHash()).setOk(rs).build();
                }
            }else {
                r = ContractProto.Response.newBuilder().setOk(false).setCodeHash(info.getCodeHash()).build();
            }
        }else  {
            r = ContractProto.Response.newBuilder().setOk(false).setCodeHash(info.getCodeHash()).build();
        }
        if (responseObserver != null) {
            responseObserver.onNext(r);
            responseObserver.onCompleted();
        }
    }

    public boolean deploy(ContractInfo info) {
        if(info.getCodeHash() != null && !info.getCodeHash().equals("")){
            String codeHash = caculateCodeHash(info.getContractPath());
            if (!codeHash.equals(info.getCodeHash())) {
                logger.error("code has been changed, origin hash: " + info.getCodeHash() + " current hash: " + codeHash);
                return false;
            }
        }
        return cm.deployContract(info);
    }

    public Task constructTask(final TaskType type, final ContractProto.Request request) {
        Task task = null;
        String cid = request.getContext().getCid(); //TODO: check this earlier
        ContractTemplate contract = cm.getContract(request.getContext().getCid());
        if (contract == null) {
            logger.warn("contract with id " + cid + " is not found!");
            return task;
        }
        switch (type) {
            case INVOKE:
                task = new InvokeTask(contract, request, constructContext(request.getContext()));
                break;
        }
        return task;
    }

    public ContractManager getContractMgr() {
        return cm;
    }

    public Context constructContext(ContractProto.RequestContext context) {
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
        String argTypes[] = new String[j - 1];
        Object objectArgs[] = new Object[j - 1];
        while (j < n) {
            String className = args.get(i).toStringUtf8();
            String arg = args.get(j).toStringUtf8();
            argTypes[i - 1] = className;
            switch (className){
                case "boolean":
                    objectArgs[i - 1] = arg == "true" ? true : false;
                    break;
                case "char":
                     objectArgs[i - 1] = arg.charAt(0);
                     break;
                case "short":
                    objectArgs[i - 1] = Short.parseShort(arg);
                    break;
                case "int":
                    objectArgs[i - 1] = Integer.parseInt(arg);
                    break;
                case "long":
                    objectArgs[i - 1] = Long.parseLong(arg);
                    break;
                case "float":
                    objectArgs[i - 1] = Float.parseFloat(arg);
                    break;
                case "double":
                    objectArgs[i - 1] = Double.parseDouble(arg);
                    break;
                case "String":
                    objectArgs[i - 1] = arg;
                    break;
                default:
                   logger.error("can not pass non string object to contract constructor");
                   return false;
            }
            i ++;
            j ++;
        }
        info.setArgTypes(argTypes);
        info.setArgs(objectArgs);
        return true;
    }

    public void caculateCodeHash(ContractInfo info) {
        String dir = info.getContractPath();
        byte[] code = IOHelper.readCode(dir);
        info.setCodeHash(HashFunction.computeCodeHash(code));
    }

    public String caculateCodeHash(String path) {
        byte[] code = IOHelper.readCode(path);
        return HashFunction.computeCodeHash(code);
    }

    public boolean checkContractExistence(ContractProto.Request request, StreamObserver<ContractProto.Response> responseObserver) {
        String cid = request.getContext().getCid();
        String err;
        ContractInfo info = cm.getContractInfoByCid(cid);
        if(info == null) {
            err = "operate contract failed, contract with id " + cid + " is nonexistent";
            logger.error(err);
            Errors.ReturnErrMsg(err, responseObserver);
            return false;
        }
        return true;
    }
}
