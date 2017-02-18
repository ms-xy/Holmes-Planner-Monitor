package org.holmesprocessing.totem.monitoring.linux;

import java.util.StringTokenizer;

public class LoadInfo {
    public double loads1;
    public double loads5;
    public double loads15;

    public void update() throws Exception {
        String fileContents = ProcFileReader.read("/proc/loadavg");

        //"0.65 1.00 1.01 1/1066 1643"
        // 1    5    15
        StringTokenizer words = new StringTokenizer(fileContents, " ", false);

        loads1 = Double.parseDouble(words.nextToken());
        loads5 = Double.parseDouble(words.nextToken());
        loads15 = Double.parseDouble(words.nextToken());
    }
}
