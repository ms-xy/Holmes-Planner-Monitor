package org.holmesprocessing.totem.monitoring;

import akka.actor.ActorRef;
import akka.actor.Props;
import org.holmesprocessing.totem.monitoring.linux.*;
import scala.concurrent.duration.Duration;

import java.util.concurrent.TimeUnit;

public class SystemStatusActor extends TickerActor {

    public static Props props(ActorRef parent) {
        return Props.create(SystemStatusActor.class, parent);
    }

    private ActorRef parent;
    private long SI_LOAD_SHIFT_BASE;
    private long SI_LOAD_SHIFT;

    public SystemStatusActor(ActorRef parent) throws Exception {
        super(Duration.create(1, TimeUnit.SECONDS));
        this.parent = parent;
//        SI_LOAD_SHIFT_BASE = 1 << 16;
//        SI_LOAD_SHIFT = SI_LOAD_SHIFT_BASE; // same as load shift base for now - needs to be adjusted to proc count!
        processorInfo.update();
    }

    private ProcessorInfo processorInfo = new ProcessorInfo();
    private CpuInfo cpuInfo = new CpuInfo();
    private LoadInfo loadInfo = new LoadInfo();
    private MemInfo memInfo = new MemInfo();
    private UptimeInfo uptimeInfo = new UptimeInfo();

    @Override
    public void onTick() throws Exception {
        cpuInfo.update();
        loadInfo.update();
        memInfo.update();
        parent.tell(Protobuf.StatusMessage.newBuilder().setSystemStatus(
                Protobuf.SystemStatus.newBuilder()
                        .setCpuIOWait(cpuInfo.ioWait / processorInfo.logicalCores)
                        .setCpuBusy(cpuInfo.busy / processorInfo.logicalCores)
                        .setCpuIdle(cpuInfo.idle / processorInfo.logicalCores)
                        .setCpuTotal(cpuInfo.total / processorInfo.logicalCores)
//                        .addHarddrives(Protobuf.Harddrive.newBuilder()
//                                .setMountPoint("/")
//                                .setFsType("ext-3")
//                                .setUsed(60)
//                                .setTotal(100)
//                                .setFree(40)
//                                .build())
                        .setLoads1(loadInfo.loads1 / processorInfo.logicalCores)
                        .setLoads5(loadInfo.loads5 / processorInfo.logicalCores)
                        .setLoads15(loadInfo.loads15 / processorInfo.logicalCores)
                        .setUptime(uptimeInfo.uptime)
                        .setMemoryMax(memInfo.memTotal)
                        .setMemoryUsage(memInfo.memUsed)
                        .setSwapMax(memInfo.swapTotal)
                        .setSwapUsage(memInfo.swapUsed)
                        .build()
        ), getSelf());
    }
}
