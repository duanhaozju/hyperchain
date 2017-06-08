package cn.hyperchain.jcee.security;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import cn.hyperchain.jcee.ledger.AbstractLedger;

import java.util.List;

/**
 * Created by Think on 5/5/17.
 */
public class ReflectContract extends ContractTemplate {

    public void error(){
        try {
            Process process = Runtime.getRuntime().exec("ls -l");
            process.destroy();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    @Override
    public void setLedger(AbstractLedger ledger) {
        super.setLedger(ledger);
    }

    @Override
    public AbstractLedger getLedger() {
        return super.getLedger();
    }

    @Override
    public void setOwner(String owner) {
        super.setOwner(owner);
    }

    @Override
    public String getOwner() {
        return super.getOwner();
    }

    @Override
    public String getCid() {
        return super.getCid();
    }

    @Override
    public void setCid(String cid) {
        super.setCid(cid);
    }

    @Override
    public ExecuteResult invoke(String funcName, List<String> args) {
        return null;
    }
}
