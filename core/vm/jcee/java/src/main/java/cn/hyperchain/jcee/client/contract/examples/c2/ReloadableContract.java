/**
 * Hyperchain License
 * Copyright (C) 2016 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.client.contract.examples.c2;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.client.contract.ContractTemplate;

import java.util.List;

/**
 * Created by wangxiaoyi on 2017/4/13.
 */
public class ReloadableContract extends ContractTemplate {

    private ImplicitClass ic;

    public ReloadableContract() {
        ic = new ImplicitClass();
    }

    @Override
    public String toString() {
        return "ImplicitClass id is " + ImplicitClass.id + " and A1 id is " + A1.id;
    }
}


class A1 {
    static int id = 2;
}
