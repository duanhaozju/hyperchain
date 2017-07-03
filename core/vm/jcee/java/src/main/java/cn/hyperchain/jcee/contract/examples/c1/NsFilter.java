package cn.hyperchain.jcee.contract.examples.c1;

import cn.hyperchain.jcee.contract.ContractInfo;
import cn.hyperchain.jcee.contract.Filter;
import org.apache.log4j.Logger;

import java.util.HashSet;

/**
 * Created by huhu on 2017/6/30.
 */
public class NsFilter implements Filter {

    private static final Logger logger = Logger.getLogger(NsFilter.class);
    private HashSet<String> nsRuler = new HashSet<>();
    @Override
    public boolean doFilter(ContractInfo info) {
        String ns = info.getNamespace();
        for(String ruler : nsRuler){
            if(ns.equals(ruler)){
                return true;
            }
        }

        return false;
    }

    public void addRuler(String ns){
        nsRuler.add(ns);
    }

    public void removeRuler(String ns){
        nsRuler.remove(ns);
    }
}
