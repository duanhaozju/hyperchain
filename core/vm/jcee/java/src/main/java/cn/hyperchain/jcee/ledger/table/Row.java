/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.ledger.table;

import com.google.gson.Gson;

import java.util.HashMap;
import java.util.Map;

/**
 * Created by wangxiaoyi on 2017/6/29.
 * @warn not thread safe
 */
public class Row {

    private String rowId; //unique identifier of a perticular row
    private Map<String, byte[]> data;

    public Row(String rowId){
        this.rowId = rowId;
        data = new HashMap<>();
    }

    public void put(String key, byte[] value){
        data.put(key, value);
    }

    public String toJSON() {
        Gson gson = new Gson();
        return gson.toJson(this);
    }

    @Override
    public String toString() {
        return "Row{" +
                "rowId='" + rowId + '\'' +
                ", data=" + data +
                '}';
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof Row)) return false;

        Row row = (Row) o;

        if (rowId != null ? !rowId.equals(row.rowId) : row.rowId != null) return false;
        return data != null ? data.equals(row.data) : row.data == null;
    }

    @Override
    public int hashCode() {
        int result = rowId != null ? rowId.hashCode() : 0;
        result = 31 * result + (data != null ? data.hashCode() : 0);
        return result;
    }
}
