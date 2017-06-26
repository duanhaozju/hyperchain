package cn.hyperchain.jcee.contract;

import java.util.HashMap;
import java.util.Map;

/**
 * Created by wangxiaoyi on 2017/6/23.
 */
public class Event {
    private String name;

    private Map<String, String> atrributes;

    public Event(String name) {
        this.name = name;
    }

    public void put(String key, String value) {
        if (atrributes == null) {
            atrributes = new HashMap<>();
        }
        atrributes.put(key, value);
    }
}
