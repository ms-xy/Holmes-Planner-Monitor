package org.holmesprocessing.totem.monitoring.linux;

import java.util.StringTokenizer;

public class CpuInfo {
    public long ioWait;
    public long idle;
    public long busy;
    public long total;

    private long prevIoWait;
    private long prevIdle;
    private long prevBusy;
    private long prevTotal;

    public void update() throws Exception {
        String fileContents = ProcFileReader.read("/proc/stat");
        StringTokenizer fileLines = new StringTokenizer(fileContents, "\n", false);
        // Pattern cpuRe = Pattern.compile("^cpu(\\d+)? +(\\d+) +(\\d+) +(\\d+)$");

        while (fileLines.hasMoreTokens()) {
            String line = fileLines.nextToken();
            if (line.startsWith("cpu ")) {
                //"cpu  28837567 164574 10602938 387346248 3905712 951 86607 0 0 0"
                //      user     nice   system   idle      iowait  irq softirq steal
                StringTokenizer words = new StringTokenizer(line, " ", false);
                words.nextToken(); // skip "cpu"

                long user = Long.parseLong(words.nextToken());
                long nice = Long.parseLong(words.nextToken());
                long system = Long.parseLong(words.nextToken());
                long idle = Long.parseLong(words.nextToken());
                long iowait = Long.parseLong(words.nextToken());
                long irq = Long.parseLong(words.nextToken());
                long softirq = Long.parseLong(words.nextToken());
                long steal = Long.parseLong(words.nextToken());
                long virtual = Long.parseLong(words.nextToken());

                long busy = user + nice + system + irq + softirq + steal;
                long total = idle + busy;

                this.ioWait = iowait - prevIoWait;
                this.idle = idle - prevIdle;
                this.busy = busy - prevBusy;
                this.total = total - prevTotal;
                // load = float64(d_total-d_idle) / float64(d_total) * 100

                prevIoWait = iowait;
                prevIdle = idle;
                prevBusy = busy;
                prevTotal = total;

                break;
            }
        }
    }
}
