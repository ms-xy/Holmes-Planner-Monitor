package org.holmesprocessing.totem.monitoring.linux;

import java.util.StringTokenizer;

public class MemInfo {
    public long memTotal;
    public long memFree;
    public long memUsed;
    public long memAvailable;
    public long buffers;
    public long cached;
    public long swapCached;
//    public long active;
//    public long inactive;
//    public long active_anon;
//    public long inactive_anon;
//    public long active_file;
//    public long inactive_file;
//    public long unevictable;
//    public long mlocked;
    public long swapTotal;
    public long swapFree;
    public long swapAvailable;
    public long swapUsed;
//    public long dirty;
//    public long writeback;
//    public long anonPages;
//    public long mapped;
    public long shmem;
//    public long slab;
    public long sReclaimable;
//    public long sUnreclaim;
//    public long kernelStack;
//    public long pageTables;
//    public long nfsUnstable;
//    public long bounce;
//    public long writebackTmp;
//    public long commitLimit;
//    public long committedAs;
//    public long vmallocTotal;
//    public long vmallocUsed;
//    public long vmallocChunk;
//    public long hardwareCorrupted;
//    public long anonHugePages;
//    public long cmaTotal;
//    public long cmaFree;
//    public long hugePagesTotal;
//    public long hugePagesFree;
//    public long hugePagesRsvd;
//    public long hugePagesSurp;
//    public long hugepagesize;
//    public long directMap4k;
//    public long direktMap2M;

    /*
    Active:          4424628 kB
    Active(anon):    3510196 kB
    Active(file):     914432 kB
    AnonPages:       4329612 kB
    AnonHugePages:   1290240 kB
    Hugepagesize:       2048 kB
    DirectMap2M:     7430144 kB
    DirectMap4k:      653060 kB
    Dirty:             11120 kB
    HardwareCorrupted:     0 kB
    HugePages_Free:        0
    HugePages_Rsvd:        0
    HugePages_Surp:        0
    HugePages_Total:       0
    Inactive:        1994096 kB
    Inactive(anon):  1354248 kB
    Inactive(file):   639848 kB
    KernelStack:       16784 kB
    NFS_Unstable:          0 kB
    PageTables:        72008 kB
    Unevictable:          64 kB
    VmallocChunk:   34359279496 kB
    VmallocTotal:   34359738367 kB
    VmallocUsed:      359876 kB
    Writeback:             0 kB
    WritebackTmp:          0 kB
    */

    public void update() throws Exception {
        String fileContents = ProcFileReader.read("/proc/meminfo");
        StringTokenizer fileLines = new StringTokenizer(fileContents, "\n", false);

        while (fileLines.hasMoreTokens()) {
            String line = fileLines.nextToken();

            StringTokenizer words = new StringTokenizer(line, " ", false);
            String name = words.nextToken();

            switch (name.charAt(0)) {
                /*
                Bounce:                0 kB
                Buffers:          272984 kB
                */
                case 'B':
                    if (name.startsWith("Buffers"))
                        buffers = Long.parseLong(words.nextToken());
                    break;
                /*
                Cached:          1792948 kB
                CmaFree:               0 kB
                CmaTotal:              0 kB
                CommitLimit:     8818868 kB
                Committed_AS:   13517460 kB
                */
                case 'C':
                    if (name.startsWith("Cached"))
                        cached = Long.parseLong(words.nextToken());
                    break;
                /*
                Mapped:           620852 kB
                MemTotal:        7872120 kB
                MemFree:         1089900 kB
                MemAvailable:    2539740 kB
                Mlocked:              64 kB
                */
                case 'M':
                    if (name.startsWith("MemTotal"))
                        memTotal = Long.parseLong(words.nextToken());
                    else if (name.startsWith("MemFree"))
                        memFree = Long.parseLong(words.nextToken());
                    else if (name.startsWith("MemAvailable"))
                        memAvailable = Long.parseLong(words.nextToken());
                    break;
                /*
                Shmem:            511652 kB
                Slab:             197172 kB
                SReclaimable:     129004 kB
                SUnreclaim:        68168 kB
                SwapCached:        67460 kB
                SwapFree:        4380880 kB
                SwapTotal:       4882808 kB
                 */
                case 'S':
                    if (name.startsWith("Shmem"))
                        shmem = Long.parseLong(words.nextToken());
                    else if (name.startsWith("SReclaimable"))
                        sReclaimable = Long.parseLong(words.nextToken());
                    else if (name.startsWith("SwapCached"))
                        swapCached = Long.parseLong(words.nextToken());
                    else if (name.startsWith("SwapFree"))
                        swapFree = Long.parseLong(words.nextToken());
                    else if (name.startsWith("SwapTotal"))
                        swapTotal = Long.parseLong(words.nextToken());
                    break;
            }
        }

        // override values and calculate others
        cached = cached + sReclaimable - shmem;
        memAvailable = memFree + buffers + cached;
        memUsed = memTotal - memAvailable;

        // apply scaling (kB)
        memTotal *= 1024;
        memFree *= 1024;
        memAvailable *= 1024;
        memUsed *= 1024;

        // override swap and apply scaling (kB)
        swapTotal = swapTotal * 1024;
        swapFree = swapFree * 1024;
        swapAvailable = swapFree * 1024;
        swapUsed = (swapTotal - swapFree) * 1024;
    }
}
