/**
 * Hyperchain License
 * Copyright (C) 2017 The Hyperchain Authors.
 */
package cn.hyperchain.jcee.server.ledger.table;

import com.google.gson.Gson;

import java.util.Arrays;
import java.util.HashMap;
import java.util.Iterator;
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
        data = new DataMap();
    }

    public String getRowId() {
        return rowId;
    }

    public void put(String key, byte[] value){
        data.put(key, value);
    }

    public String get(String key) {
        return new String(data.get(key));
    }

    public String toJSON() {
        Gson gson = new Gson();
        return gson.toJson(this);
    }

    public Row merge(Row anotherRow) {
        this.data.putAll(anotherRow.data);
        return this;
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

    //self define equals for map value with type of byte[]
    class DataMap extends HashMap<String, byte[]> {

        @Override
        public boolean equals(Object o) {
            if (o == this)
                return true;

            if (!(o instanceof Map))
                return false;
            Map<?,?> m = (Map<?,?>) o;
            if (m.size() != size())
                return false;

            try {
                Iterator<Entry<String,byte[]>> i = entrySet().iterator();
                while (i.hasNext()) {
                    Entry<String, byte[]> e = i.next();
                    String key = e.getKey();
                    byte[] value = e.getValue();
                    if (value == null) {
                        if (!(m.get(key)==null && m.containsKey(key)))
                            return false;
                    } else {
                        if (!Arrays.equals(value, (byte[]) m.get(key))) {
                            return false;
                        }
                    }
                }
            } catch (ClassCastException unused) {
                return false;
            } catch (NullPointerException unused) {
                return false;
            }

            return true;
        }
    }
}
