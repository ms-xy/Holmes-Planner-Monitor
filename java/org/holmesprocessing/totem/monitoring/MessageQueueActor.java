package org.holmesprocessing.totem.monitoring;

import akka.actor.Actor;
import akka.actor.ActorRef;
import akka.actor.Props;
import akka.actor.UntypedActor;
import akka.actor.dsl.Creators;
import scala.concurrent.duration.Duration;

import java.util.Queue;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.concurrent.TimeUnit;

public class MessageQueueActor extends UntypedActor {

    public static Props props(ActorRef parent) {
        return Props.create(MessageQueueActor.class, parent);
    }

    private ActorRef parent;
    private Queue<Tuple> queue;

    public MessageQueueActor(ActorRef parent) {
        this.parent = parent;
        queue = new ConcurrentLinkedQueue<>();
    }

    public static final class Tuple {
        public Object message;
        public ActorRef sender;
        public Tuple(Object o, ActorRef a) {
            message = o;
            sender = a;
        }
    }

    public static final class Message {}
    public static final Message Start = new Message();
    private boolean run = false;

    @Override
    public void onReceive(Object message) throws Exception {
        if (run) {
            parent.tell(message, getSelf());
            return;
        }
        if (message.equals(Start)) {
            run = true;
            queue.forEach(o -> parent.tell(o.message, o.sender));
            queue = null;
        } else {
            queue.add(new Tuple(message, getSender()));
        }
    }
}
