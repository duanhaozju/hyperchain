package cn.hyperchain.jcee.contract.security;

import jdk.internal.org.objectweb.asm.Opcodes;

import java.util.HashMap;
import java.util.Map;

/**
 * Created by Think on 4/26/17.
 */
public class Rule {
    public static final Map<String, String> allowedRules = new HashMap();
    public static final Map<String, String> notAllowedRules = new HashMap();
    public static final Map<String, String> notAllowedIORules = new HashMap();
    public static final Map<String, String> notAllowedSysCallRules = new HashMap();
    public static final Map<String, String> notAllowedThreadRules = new HashMap();
    public static final Map<String, String> notAllowedNetRules = new HashMap();
    public static final Map<String, String> notAllowedReflectRules = new HashMap();

    private static final String descIO = "not allow io operation.";
    private static final String descSysCall = "not allow system call operation.";
    private static final String descThread = "not allow thread(concurrent) operation.";
    private static final String descNet = "not allow net operation.";
    private static final String descReflect = "not allow reflect operation.";

    static {
        allowedRules.put(String.valueOf(Opcodes.ACC_FINAL), "allow final type member variables.");

        notAllowedIORules.put("io", descIO);
        notAllowedIORules.put("Readable", descIO);
        notAllowedIORules.put("nio", descIO);

        notAllowedSysCallRules.put("Process", descSysCall);
        notAllowedSysCallRules.put("ProcessBuilder", descSysCall);
        notAllowedSysCallRules.put("ProcessBuilder$Redirect", descSysCall);
        notAllowedSysCallRules.put("Runtime", descSysCall);

        notAllowedSysCallRules.put("Thread$State", descThread);
        notAllowedThreadRules.put("Thread", descThread);
        notAllowedThreadRules.put("Runnable", descThread);
        notAllowedThreadRules.put("ThreadGroup", descThread);
        notAllowedThreadRules.put("ThreadLocal", descThread);
        notAllowedThreadRules.put("concurrent", descThread);
        notAllowedThreadRules.put(String.valueOf(Opcodes.ACC_SYNCHRONIZED), descThread);

        notAllowedNetRules.put("net", descNet);

        notAllowedReflectRules.put("Class", descReflect);

        notAllowedRules.putAll(notAllowedIORules);
        notAllowedRules.putAll(notAllowedSysCallRules);
        notAllowedRules.putAll(notAllowedThreadRules);
        notAllowedRules.putAll(notAllowedNetRules);
        notAllowedRules.putAll(notAllowedReflectRules);
    }
}
