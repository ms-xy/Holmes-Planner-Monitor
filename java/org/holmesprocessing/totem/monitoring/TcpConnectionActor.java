package org.holmesprocessing.totem.monitoring;

import akka.actor.ActorRef;
import akka.actor.Props;
import akka.actor.UntypedActor;
import akka.io.Tcp;
import akka.io.TcpMessage;
import akka.io.UdpConnected;
import akka.io.UdpConnectedMessage;
import akka.japi.Procedure;
import akka.util.ByteString;

import java.net.InetSocketAddress;

import static org.holmesprocessing.totem.monitoring.TcpConnectionActor.ConnectionState.*;

public class TcpConnectionActor extends UntypedActor {


    public static final class MessageClass {
    }

    public static final MessageClass Disconnected = new MessageClass();
    public static final MessageClass Connected = new MessageClass();

    public enum ConnectionState {
        _disconnected,
        _connected
    }

    private ConnectionState state;

    private final ActorRef tcp;
    private final ActorRef callback;

    public static Props props(InetSocketAddress remoteAddress, ActorRef parent) {
        return Props.create(TcpConnectionActor.class, remoteAddress, parent);
    }

    public TcpConnectionActor(InetSocketAddress remoteAddress, ActorRef callback) {
        this.tcp = Tcp.get(getContext().system()).manager();
        tcp.tell(TcpMessage.connect(remoteAddress), getSelf());
        this.callback = callback;
    }

    @Override
    public void onReceive(Object msg) throws Exception {
        switch (this.state) {
            case _connected:
                stateConnected(msg);
                break;
            case _disconnected:
                stateDisconnected(msg);
                break;
        }
    }

    public void stateDisconnected(Object msg) throws Exception {
        if (msg instanceof Tcp.Connected) {
            state = _connected;
            callback.tell(Connected, getSelf());

        } else if (msg instanceof Tcp.CommandFailed) {
            callback.tell(Disconnected, getSelf());

        } else {
            unhandled(msg);
        }
    }

    public void stateConnected(Object msg) throws Exception {
        if (msg instanceof Tcp.Received) {
            Tcp.Received r = (Tcp.Received) msg;
            Protobuf.ControlMessage cm = Protobuf.ControlMessage.parseFrom(r.data().toArray());
            callback.tell(cm, getSelf());

        } else if (msg instanceof Protobuf.StatusMessage) {
            Protobuf.StatusMessage m = (Protobuf.StatusMessage) msg;
            byte[] data = m.toByteArray();
            tcp.tell(UdpConnectedMessage.send(ByteString.fromArray(data)), getSelf());

        } else if (msg instanceof Tcp.CommandFailed) {
            // TODO: what to do in case of failed command?
            // callback.tell(msg, getSelf());

        } else if (msg instanceof  Tcp.ConnectionClosed) {
            System.out.println("[tcp] disconnected: " + msg.toString() + " -- " + msg.getClass().toString());
            state = _disconnected;
            callback.tell(Disconnected, getSelf());

        } else {
            System.out.println("[tcp] unhandled: " + msg.toString() + " -- " + msg.getClass().toString());
            unhandled(msg);
        }
    }

    @Override
    public void postStop() throws Exception {
        super.postStop();
    }
}
