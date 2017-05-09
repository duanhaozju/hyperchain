package cn.hyperchain.jcee.contract.examples.c2;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import com.google.protobuf.ByteString;

import java.util.List;

/**
 * Created by wangxiaoyi on 2017/4/13.
 */
public class ReloadableContract extends ContractTemplate {

    private ImplicitClass ic;

    public ReloadableContract() {
        ic = new ImplicitClass();
    }

    /**
     * invoke smart contract method
     *
     * @param funcName function name user defined in contract
     * @param args     arguments of funcName
     */
    @Override
    public ExecuteResult invoke(String funcName, List<String> args) {
        return result(false);
    }

    @Override
    public String toString() {
        return "ImplicitClass id is " + ImplicitClass.id + " and A1 id is " + A1.id;
    }
}


class A1 {
    static int id = 2;
}
