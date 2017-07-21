/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.contract.examples.c1;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import cn.hyperchain.jcee.contract.filter.FilterChain;

import java.util.List;

/**
 * Created by wangxiaoyi on 2017/4/7.
 */
public class MySmartContract extends ContractTemplate {

    public ExecuteResult invoke(String funcName, List<String> args) {
        //logger.info("invoke function name: " + funcName);
        switch (funcName) {
            case "test": {
                this.test("invoke method:" + args.get(0));
                return result(true);
            }
            case "authority":
                return authority(args);
            case "addNsRuler":
                return addNsRuler(args);
            case "addCidRuler":
                return addCidRuler(args);
            default:
                logger.error("no such method found");
        }

        return result(false);
    }

    /**
     * openInvoke provide a interface to be invoked by other contracts
     *
     * @param funcName contract name
     * @param args     arguments
     * @return {@link ExecuteResult}
     */
    @Override
    protected ExecuteResult openInvoke(String funcName, List<String> args) {
        switch (funcName) {
            case "test": {
                this.test("invoke method:" + args.get(0));
                return result(true);
            }
            default:
                logger.error("no such method found or the method cannot be invoked by other contract");
        }
        return result(false);
    }

    public void test(String name) {
        logger.info("invoke in test");
    }


    public ExecuteResult authority(List<String> args){
        logger.info("no authority");
        return result(true);
    }

    public void init(){
//        FilterChain fc = this.getFilterChain();
//
//        NsFilter nsFilter = new NsFilter();
//        nsFilter.addRuler("global");
//
//        CidFilter cidFilter = new CidFilter();
//        cidFilter.addRuler("bbe2b6412ccf633222374de8958f2acc76cda9c9");
//
//        //fc.addFilter("namespace",nsFilter);
//        //fc.addFilter("cid",cidFilter);

    }

    public ExecuteResult addNsRuler(List<String> args){
//        FilterChain fc = this.getFilterChain();
//       // NsFilter nsFilter = (NsFilter)fc.getFilter("namespace");
//        for(String ns:args){
//         //   nsFilter.addRuler(ns);
//        }
//        return result(true);
        return result(true);
    }

    public ExecuteResult addCidRuler(List<String> args){
//        FilterChain fc = this.getFilterChain();
//      //  CidFilter cidFilter = (CidFilter) fc.getFilter("cid");
//        for(String cid:args){
//        //    cidFilter.addRuler(cid);
//        }
//        return result(true);
        return result(true);
    }
}
