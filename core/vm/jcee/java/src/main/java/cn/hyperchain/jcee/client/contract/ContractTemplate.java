/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.client.contract;

import cn.hyperchain.jcee.client.contract.filter.Filter;
import cn.hyperchain.jcee.client.contract.filter.FilterChain;
import cn.hyperchain.jcee.client.contract.filter.FilterManager;
import cn.hyperchain.jcee.client.executor.AbstractContractHandler;
import cn.hyperchain.jcee.client.executor.IHandler;
import cn.hyperchain.jcee.client.ledger.table.RelationDB;
import cn.hyperchain.jcee.client.ledger.table.Table;
import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.common.Context;
import cn.hyperchain.jcee.client.ledger.AbstractLedger;
import cn.hyperchain.jcee.client.ledger.Result;
import cn.hyperchain.jcee.client.ledger.table.TableName;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;
import org.apache.log4j.Logger;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.util.List;

//ContractBase which is used as a skeleton of smart contract
@AllArgsConstructor
@NoArgsConstructor
public class ContractTemplate {
    @Setter
    @Getter
    private ContractInfo info;
    @Setter
    @Getter
    private String owner;
    @Setter
    @Getter
    private String cid;
    @Setter
    @Getter
    protected AbstractLedger ledger;
    protected Logger logger = Logger.getLogger(this.getClass().getSimpleName());
    protected final FilterManager filterManager = new FilterManager();

    protected final ExecuteResult sysQuery(QueryType type) {
        switch (type) {
            case CONTRACT_INFO:
                return result(true, info.toString());
            case DATABASE_SCHEMAS:
                return getDBSchema();
            default:
                return result(false, "query type " + type + "not found");
        }
    }

    /**
     * invoke smart contract method
     *
     * @param funcName function name user defined in contract
     * @param args     arguments of funcName
     * @return {@link ExecuteResult}
     */
    public final ExecuteResult invoke(String funcName, List<String> args) {
        Class clazz = this.getClass();
        try {
            Method method = clazz.getDeclaredMethod(funcName, List.class);
            if(!method.isAccessible()) {
                method.setAccessible(true);
            }
            return (ExecuteResult) method.invoke(this, args);
        } catch (NoSuchMethodException e) {
            logger.error("no such method");
        } catch (IllegalAccessException e) {
            logger.error(e.getMessage());
        } catch (InvocationTargetException e) {
            logger.error(e.getMessage());
        } catch (Exception e) {
            logger.error(e.getMessage());
        }
        return result(false, "Contract invoke error");
    }

    /**
     * openInvoke provide a interface to be invoked by other contracts
     * @param funcName contract name
     * @param args arguments
     * @return {@link ExecuteResult}
     */
    protected ExecuteResult openInvoke(String funcName, List<String> args){
        //this is a empty method by default
        return result(false, "no method to invoke");
    }

    /**
     * invoke other contract method, this method is not safe enough now
     * @param ns contract namespace
     * @param contractAddr contract address
     * @param func function name
     * @param args function arguments
     * @return {@link ExecuteResult}
     */
    public final ExecuteResult invokeContract(String ns, String contractAddr, String func, List<String> args) {
        //TODO: how to do the authority control

        // 1.check invoke identifier namespace and the contract Address
        if (ns == null || ns.isEmpty() || contractAddr == null || contractAddr.isEmpty())
            return result(false, "invalid namespace or contract address");

        AbstractContractHandler ch = AbstractContractHandler.getContractHandler();
        IHandler handler = ch.get(ns);

        if(! ch.hasHandlerForNamespace(ns)) {
            return result(false, String.format("no namespace named: %s found", ns));
        }
        ContractTemplate ct = handler.getContractMgr().getContract(contractAddr);
        if (ct == null) {
            return result(false, String.format("no contract with address: %s found", contractAddr));
        }

        Context context = new Context(ledger.getContext().getId());
        context.setRequestContext(ledger.getContext().getRequestContext());
        return ct.securityCheck(func, args, context);
    }

    private final ExecuteResult securityCheck(String funcName, List<String> args, Context context){

        FilterChain fc = filterManager.getFilterChain(funcName);
        if (fc != null && !fc.doFilter(context)) {
            return result(false,
                    String.format("Invoker %s at contract %s try to invoke function %s failed", context.getRequestContext().getInvoker()
                    , context.getRequestContext().getCid(), funcName));
        }
        return openInvoke(funcName, args);
    }

    /**
     * execution result encapsulation method.
     * @param exeSuccess indicate the execute status, true for success, false for failed
     * @param result contract execution result if success or failed message
     * @return
     */
    protected final ExecuteResult result(boolean exeSuccess, Object result) {
        ExecuteResult rs = new ExecuteResult<>(exeSuccess, result);
        return rs;
    }

    //just indicate the execution status, no result specified
    protected final ExecuteResult result(boolean exeSuccess) {
        ExecuteResult rs = new ExecuteResult<>(exeSuccess, "");
        return rs;
    }

    protected final void addFilter(String funcName, Filter filter) {
        filterManager.AddFilter(funcName, filter);
    }

    public String getNamespace() {
        return this.info.getNamespace();
    }
    /**
     * get table instance used by this contract with specified name
     * @param name name
     * @return Table instance
     */
    public Table getTable(String name) {
        RelationDB db = ledger.getDataBase();
        return db.getTable(new TableName(getNamespace(), getCid(), name));
    }

    public ExecuteResult getDBSchema() {
        Result rs = ledger.get("database_schemas");
        if (rs.isEmpty()) {
            return result(false, "Database schemas are empty");
        } else {
            return result(true, rs.toString());
        }
    }

    public enum QueryType {
        CONTRACT_INFO,
        DATABASE_SCHEMAS
    }
}