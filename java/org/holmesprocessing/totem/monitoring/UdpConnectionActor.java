package org.holmesprocessing.totem.monitoring;

import akka.actor.ActorRef;
import akka.actor.Props;
import akka.actor.UntypedActor;
import akka.io.UdpConnected;
import akka.io.UdpConnectedMessage;
import akka.japi.Procedure;
import akka.util.ByteString;

import java.net.InetSocketAddress;

public class UdpConnectionActor extends UntypedActor {

    public static Props props(InetSocketAddress remoteAddress, ActorRef parent) {
        return Props.create(UdpConnectionActor.class, remoteAddress, parent);
    }

    public UdpConnectionActor(InetSocketAddress remoteAddress, ActorRef parent) {
        ActorRef udp = UdpConnected.get(getContext().system()).manager();
        udp.tell(UdpConnectedMessage.connect(getSelf(), remoteAddress), getSelf());
        this.parent = parent;
    }


    private final ActorRef parent;

    public static final class ReadyClass {
    }

    public static final ReadyClass Ready = new ReadyClass();


    @Override
    public void onReceive(Object msg) throws Exception {
        if (msg instanceof UdpConnected.Connected) {
            getContext().become(ready(getSender(), parent));
            parent.tell(Ready, getSelf());
        } else unhandled(msg);
    }

    private Procedure<Object> ready(final ActorRef connection, final ActorRef parent) {
        return new Procedure<Object>() {
            @Override
            public void apply(Object msg) throws Exception {
                if (msg instanceof UdpConnected.Received) {
                    final UdpConnected.Received r = (UdpConnected.Received) msg;
                    // process data, send it on, etc.
                    Protobuf.ControlMessage cm = Protobuf.ControlMessage.parseFrom(r.data().toArray());
                    parent.tell(cm, getSelf());
                } else if (msg instanceof Protobuf.StatusMessage) {
                    Protobuf.StatusMessage m = (Protobuf.StatusMessage) msg;
                    byte[] data = m.toByteArray();
                    connection.tell(UdpConnectedMessage.send(ByteString.fromArray(data)), getSelf());
                } else if (msg instanceof UdpConnected.CommandFailed) {
                    UdpConnected.CommandFailed cmd = (UdpConnected.CommandFailed) msg;
                    UdpConnected.Command c = cmd.cmd();
                    connection.tell(cmd, getSender());
                } else if (msg.equals(UdpConnectedMessage.disconnect())) {
                    System.out.println("disconnect received: " + msg.toString() + " -- " + msg.getClass().toString());
                    connection.tell(msg, getSelf());
                } else if (msg instanceof UdpConnected.Disconnected) {
                    System.out.println("udp module is disconnecting!");
                    getContext().stop(getSelf());
                } else {
                    System.out.println("uncaught message: " + msg.toString() + " -- " + msg.getClass().toString());
                    unhandled(msg);
                }
            }
        };
    }

    @Override
    public void postStop() throws Exception {
        super.postStop();
    }
}
