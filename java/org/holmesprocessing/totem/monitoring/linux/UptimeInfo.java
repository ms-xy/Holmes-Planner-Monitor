package org.holmesprocessing.totem.monitoring.linux;

import java.util.StringTokenizer;

public class UptimeInfo {
    public long uptime;
    public long idletime;

    public void update() throws Exception {
        String fileContents = ProcFileReader.read("/proc/uptime");

        // 5121856.15 15694030.55
        // uptime     idletime     both in seconds
        StringTokenizer words = new StringTokenizer(fileContents, " ", false);

        uptime = (long) Double.parseDouble(words.nextToken());
        idletime = (long) Double.parseDouble(words.nextToken());
    }
}
