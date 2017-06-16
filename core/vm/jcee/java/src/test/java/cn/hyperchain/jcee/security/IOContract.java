package cn.hyperchain.jcee.security;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import cn.hyperchain.jcee.ledger.AbstractLedger;

import java.io.File;
import java.util.List;

/**
 * Created by Think on 4/27/17.
 */
public class IOContract extends ContractTemplate {

    public void send(){
        File file = new File("");
    }

    public IOContract() {
        super();
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
