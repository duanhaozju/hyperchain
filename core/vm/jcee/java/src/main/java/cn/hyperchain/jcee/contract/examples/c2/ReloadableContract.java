package cn.hyperchain.jcee.contract.examples.c2;

import cn.hyperchain.jcee.contract.ContractBase;
import com.google.protobuf.ByteString;

import java.util.List;

/**
 * Created by wangxiaoyi on 2017/4/13.
 */
public class ReloadableContract extends ContractBase {

    private ImplicitClass ic;

    public ReloadableContract() {
        ic = new ImplicitClass();
    }

    /**
     * Invoke smart contract method
     *
     * @param funcName function name user defined in contract
     * @param args     arguments of funcName
     */
    @Override
    public boolean Invoke(String funcName, List<String> args) {
        return false;
    }

    /**
     * Query data stored in the smart contract
     *
     * @param funcName function name
     * @param args     function related arguments
     * @return the query result
     */
    @Override
    public ByteString Query(String funcName, List<String> args) {
        return null;
    }

    @Override
    public String toString() {
        return "ImplicitClass id is " + ImplicitClass.id + " and A1 id is " + A1.id;
    }
}


class A1 {
    static int id = 2;
}
