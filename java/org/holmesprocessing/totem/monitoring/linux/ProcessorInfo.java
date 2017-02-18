package org.holmesprocessing.totem.monitoring.linux;

import java.util.StringTokenizer;

public class ProcessorInfo {
    public long logicalCores;
    public long physicalCores;

    public void update() throws Exception {
        String fileContents = ProcFileReader.read("/proc/cpuinfo");

        /*
        processor       : 0
        vendor_id       : AuthenticAMD
        cpu family      : 16
        model           : 4
        model name      : AMD Phenom(tm) II X4 B95 Processor
        stepping        : 2
        cpu MHz         : 800.000
        cache size      : 512 KB
        physical id     : 0
        siblings        : 4
        core id         : 0
        cpu cores       : 4
        apicid          : 0
        fpu             : yes
        fpu_exception   : yes
        cpuid level     : 5
        wp              : yes
        flags           : fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx fxsr sse sse2 ht syscall nx mmxext fxsr_opt pdpe1gb rdtscp lm 3dnowext 3dnow constant_tsc nonstop_tsc pni cx16 popcnt lahf_lm cmp_legacy svm extapic cr8_legacy altmovcr8 abm sse4a misalignsse 3dnowprefetch osvw
        bogomips        : 5984.99
        TLB size        : 1024 4K pages
        clflush size    : 64
        cache_alignment : 64
        address sizes   : 48 bits physical, 48 bits virtual
        power management: ts ttp tm stc 100mhzsteps hwpstate [8]
         */
        StringTokenizer lines = new StringTokenizer(fileContents, "\n", false);

        while (lines.hasMoreTokens()) {
            String line = lines.nextToken();
            StringTokenizer words = new StringTokenizer(line, " ", false);
            String key = words.nextToken();
            System.out.println(line);
            System.out.println(key);
            if (key.startsWith("processor")) {
                logicalCores = Math.max(logicalCores, Long.parseLong(words.nextToken()));

            } else if (key.startsWith("physical")) {
                words.nextToken(); // skip colon
                physicalCores = Math.max(physicalCores, Long.parseLong(words.nextToken()));
            }
        }

        // adjust as indexes start by 0
        physicalCores += 1;
        logicalCores += 1;

        System.out.println("physical cores: " + physicalCores + " /  logical cores: " + logicalCores);
    }
}
