package cn.hyperchain.jcee.contract;


import com.google.gson.Gson;
import lombok.Getter;

import java.util.*;

/**
 * Created by wangxiaoyi on 2017/6/23.
 */
public class Event {
    @Getter
    private String name;

    private Map<String, String> atrributes;

    @Getter
    private Set<String> topics;

    public Event(String name) {
        this.name = name;
    }

    public Event(String name, String... topics) {
        this(name);
        this.topics = new HashSet<>();
        for (String topic: topics) {
            this.topics.add(topic);
        }
    }

    public void addTopic(String topic) {
        if (topics == null) {
            topics = new HashSet<>();
        }
        topics.add(topic);
    }

    public void put(String key, String value) {
        if (atrributes == null) {
            atrributes = new HashMap<>();
        }
        atrributes.put(key, value);
    }

    private String toJsonString() {
        Gson gson = new Gson();
        return gson.toJson(this);
    }

    @Override
    public String toString() {
        return toJsonString();
    }
}
