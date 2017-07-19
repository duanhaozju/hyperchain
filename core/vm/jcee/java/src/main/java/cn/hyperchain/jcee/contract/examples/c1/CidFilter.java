package cn.hyperchain.jcee.contract.examples.c1;

import cn.hyperchain.jcee.contract.ContractInfo;
import cn.hyperchain.jcee.contract.filter.Filter;
import cn.hyperchain.jcee.executor.Context;
import org.apache.log4j.Logger;

import java.util.HashSet;

/**
 * Created by huhu on 2017/7/3.
 */
public class CidFilter implements Filter {

    protected final Logger logger = Logger.getLogger(CidFilter.class);
    private HashSet<String> cidRuler = new HashSet<>();
//    @Override
//    public boolean doFilter(ContractInfo info) {
//        String cid = info.getCid();
//
//        for(String rule : cidRuler){
//            if(cid.equals(rule)){
//                return true;
//            }
//        }
//        return false;
//    }


    @Override
    public boolean doFilter(Context context) {
        return false;
    }

    @Override
    public String getName() {
        return null;
    }

    public void addRuler(String ns){
        cidRuler.add(ns);
    }

    public void removeRuler(String ns){
        cidRuler.remove(ns);
    }
}
