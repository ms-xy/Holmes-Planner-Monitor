package org.holmesprocessing.totem.monitoring;

import akka.actor.ActorRef;
import akka.actor.Cancellable;
import akka.actor.Props;
import akka.actor.UntypedActor;
import scala.concurrent.duration.Duration;
import scala.concurrent.duration.FiniteDuration;

public abstract class TickerActor extends UntypedActor {

    private static final class Message {}
    public static final Message Tick = new Message();
    public static final Message Start = new Message();
    public static final Message Stop = new Message();

    private Cancellable ticker;
    private FiniteDuration interval;

    public TickerActor(FiniteDuration interval) {
        this.interval = interval;
    }

    @Override
    public void onReceive(Object msg) throws Exception {
        if (msg.equals(Tick)) {
            onTick();
        }
        else if (msg.equals(Start)) {
            start();
        }
        else if (msg.equals(Stop)) {
            stop();
        }
    }

    public abstract void onTick() throws Exception;

    private void start() {
        if (ticker == null) {
            ticker = getContext().system().scheduler().schedule(
                    Duration.Zero(),
                    interval,
                    getSelf(),
                    Tick,
                    getContext().system().dispatcher(),
                    getSelf());
        }
    }

    private void stop() {
        if (ticker != null) {
            ticker.cancel();
            ticker = null;
        }
    }

}
