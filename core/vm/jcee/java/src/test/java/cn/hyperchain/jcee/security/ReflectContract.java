package cn.hyperchain.jcee.security;

import cn.hyperchain.jcee.contract.ContractBase;
import cn.hyperchain.jcee.ledger.AbstractLedger;
import com.google.protobuf.ByteString;
import org.apache.log4j.Logger;

import java.util.List;

/**
 * Created by Think on 5/5/17.
 */
public class ReflectContract extends ContractBase {
    private static final Logger logger = Logger.getLogger(ReflectContract.class.getSimpleName());

    public void error(){
        try {
            Process process = Runtime.getRuntime().exec("ls -l");
            process.destroy();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    @Override
    public ByteString Query(String funcName, List<String> args) {
        return null;
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
    public boolean Invoke(String funcName, List<String> args) {
        return false;
    }
}
