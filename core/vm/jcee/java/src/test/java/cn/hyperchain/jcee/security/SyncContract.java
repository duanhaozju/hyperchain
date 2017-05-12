package cn.hyperchain.jcee.security;

import cn.hyperchain.jcee.common.ExecuteResult;
import cn.hyperchain.jcee.contract.ContractTemplate;
import cn.hyperchain.jcee.ledger.AbstractLedger;
import com.google.protobuf.ByteString;
import org.apache.log4j.Logger;

import java.util.List;

/**
 * Created by Think on 4/27/17.
 */
public class SyncContract extends ContractTemplate {
    private static final Logger logger = Logger.getLogger(SyncContract.class.getSimpleName());

    public synchronized void send(){

    }

    public SyncContract() {
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
