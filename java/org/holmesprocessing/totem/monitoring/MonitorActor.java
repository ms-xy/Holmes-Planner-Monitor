package org.holmesprocessing.totem.monitoring;

import java.io.IOException;
import java.net.*;
import java.nio.ByteBuffer;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.Iterator;
import java.util.Spliterator;
import java.util.UUID;
import java.util.function.Consumer;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

// monitor specific imports
import akka.actor.*;
import akka.io.UdpConnectedMessage;
import akka.japi.Procedure;
import com.google.protobuf.ByteString;


public class MonitorActor extends UntypedActor {

    private static ActorRef singleton;

    public static void CreateInstance(ActorSystem system) {
        if (singleton == null) {
            singleton = system.actorOf(MonitorActor.props());
        }
    }

    public static void Connect(String address, String ownName, String ownAddress) {
        if (singleton == null) throw new NullPointerException();
        singleton.tell(new MsgConnect(address, ownName, ownAddress, null), singleton);
    }

    public static void PublishConfiguration(String configName) throws NullPointerException {
        if (singleton == null) throw new NullPointerException();
        singleton.tell(Protobuf.StatusMessage.newBuilder().setPlannerStatus(
                Protobuf.PlannerStatus.newBuilder()
                        .setConfigProfileName(configName)
                        .build()
        ), singleton);
    }

    public static void PublishLogs(String[] logs) throws NullPointerException {
        if (singleton == null) throw new NullPointerException();
        singleton.tell(Protobuf.StatusMessage.newBuilder().setPlannerStatus(
                Protobuf.PlannerStatus.newBuilder()
                        .addAllLogs(new StringListIterator(logs))
                        .build()
        ), singleton);
    }

    public static void PublishService(int port, String name, String configName) {
        if (singleton == null) throw new NullPointerException();
        singleton.tell(Protobuf.StatusMessage.newBuilder().setServiceStatus(
                Protobuf.ServiceStatus.newBuilder()
                        .setPort(port)
                        .setName(name)
                        .setConfigProfileName(configName)
                        .build()
        ), singleton);
    }

    public static void PublishServiceLogs(int port, String[] logs) {
        if (singleton == null) throw new NullPointerException();
        singleton.tell(Protobuf.StatusMessage.newBuilder().setServiceStatus(
                Protobuf.ServiceStatus.newBuilder()
                        .setPort(port)
                        .addAllLogs(new StringListIterator(logs))
                        .build()
        ), singleton);
    }

    public static Props props() {
        return Props.create(MonitorActor.class);
    }

    public static final class MsgConnect {
        private final String remote;
        private final String name;
        private final String address;
        private final ActorRef parent;

        public MsgConnect(String statusAddr, String plannerName, String plannerAddr, ActorRef parent) {
            remote = statusAddr;
            name = plannerName;
            address = plannerAddr;
            this.parent = parent;
        }
    }

    public static final class DisconnectClass {
    }

    public static final DisconnectClass MsgDisconnect = new DisconnectClass();

    private MonitorActorState monitorActorState;
    private UUID uuid;
    private UUID machineUuid;
    private String machineUuidFilePath;
    private ActorRef connection;
    private ActorRef systemStatusActor;
    private ActorRef parent;

    private String remoteAddress;
    private String plannerName;
    private String plannerAddress;

    private ActorRef queue;

    public MonitorActor() throws SocketException {
        // TODO: logger
        // TODO: make the machineUuid location configurable

        // Init logger (with prefix)

        // Init monitor
        uuid = new UUID(0, 0);
        machineUuid = new UUID(0, 0);
        machineUuidFilePath = "/var/tmp/holmes_processing_cache/uuid";
        systemStatusActor = getContext().actorOf(SystemStatusActor.props(getSelf()));

        // Init monitorActorState manager
        monitorActorState = new MonitorActorState(MonitorActorState.ConnectionState.Disconnected);
        queue = getContext().actorOf(MessageQueueActor.props(getSelf()));
    }

    @Override
    public void preStart() {
        getContext().become(disconnected);
    }

    @Override
    public void onReceive(Object msg) throws Exception {
    }


    private Procedure<Object> disconnected = new Procedure<Object>() {
        @Override
        public void apply(Object msg) throws Exception {
            if (msg instanceof MsgConnect) {
                MsgConnect c = (MsgConnect) msg;
                parent = c.parent;
                remoteAddress = c.remote;
                plannerName = c.name;
                plannerAddress = c.address;
                InetSocketAddress remoteAddr = parseAddress(remoteAddress);
                // Switch context
                getContext().become(connecting);
                monitorActorState.set(MonitorActorState.ConnectionState.Connecting);
                // Trigger connection
                connection = getContext().actorOf(UdpConnectionActor.props(remoteAddr, getSelf()));
                // Allow message queueing
                queue = getContext().actorOf(MessageQueueActor.props(getSelf()));
            } else {
                unhandled(msg);
            }
        }
    };

    private Procedure<Object> connecting = new Procedure<Object>() {
        @Override
        public void apply(Object msg) throws Exception {
            if (msg instanceof Protobuf.StatusMessage.Builder) {
                secured_send((Protobuf.StatusMessage.Builder) msg);
            }
            else if (msg.equals(UdpConnectionActor.Ready)) {
                // Switch context
                getContext().become(handshaking);
                monitorActorState.set(MonitorActorState.ConnectionState.ExpectedHandshakeAck);
                // Trigger handshake
                loadMachineUuid();
                send(Protobuf.StatusMessage.newBuilder().setPlannerInfo(
                        Protobuf.PlannerInfo.newBuilder()
                                .setConnect(true)
                                .setName(plannerName)
                                .setListenAddress(plannerAddress)
                ));
            } else if (msg.equals(MsgDisconnect)) {
                disconnect();
            } else {
                unhandled(msg);
            }
        }
    };

    private Procedure<Object> handshaking = new Procedure<Object>() {
        @Override
        public void apply(Object msg) throws Exception {
            if (msg instanceof Protobuf.StatusMessage.Builder) {
                secured_send((Protobuf.StatusMessage.Builder) msg);
            }
            else if (msg instanceof Protobuf.ControlMessage) {
                Protobuf.ControlMessage cm = (Protobuf.ControlMessage) msg;
                if (cm.getAckConnect()) {
                    uuid = uuidFromBytes(cm.getUuid().toByteArray());
                    machineUuid = uuidFromBytes(cm.getMachineUuid().toByteArray());
                    // Switch context
                    getContext().become(connected);
                    monitorActorState.set(MonitorActorState.ConnectionState.Connected);
                    // Save uuid
                    System.out.println("uuid / machine_uuid: " + uuid.toString() + " / " + machineUuid.toString());
                    saveMachineUuid();
                    // Periodically fetch SystemStatus
                    systemStatusActor.tell(SystemStatusActor.Start, getSelf());
                    // Start working down the accumulated message queue
                    queue.tell(MessageQueueActor.Start, getSelf());
                }
            } else if (msg.equals(MsgDisconnect)) {
                disconnect();
            } else {
                unhandled(msg);
            }
        }
    };

    private Procedure<Object> connected = new Procedure<Object>() {
        @Override
        public void apply(Object msg) throws Exception {
            if (msg instanceof Protobuf.StatusMessage.Builder) {
                secured_send((Protobuf.StatusMessage.Builder) msg);
            }
            else if (msg instanceof Protobuf.ControlMessage) {
                Protobuf.ControlMessage cm = (Protobuf.ControlMessage) msg;
                if (parent != null) {
                    parent.tell(cm, getSelf());
                }
            }
            else if (msg.equals(MsgDisconnect)) {
                disconnect();
            }
            else {
                unhandled(msg);
            }
        }
    };

    private void disconnect() {
        // Stop fetching SystemStatus
        systemStatusActor.tell(SystemStatusActor.Stop, getSelf());
        // Send disconnect message, then destroy connection
        send(Protobuf.StatusMessage.newBuilder().setPlannerInfo(
                Protobuf.PlannerInfo.newBuilder()
                        .setDisconnect(true)
                        .build()
        ));
        queue = null;
        connection.tell(UdpConnectedMessage.disconnect(), getSelf());
        connection.tell(PoisonPill.getInstance(), getSelf());
        connection = null;
        // Switch context
        getContext().become(disconnected);
        monitorActorState.set(MonitorActorState.ConnectionState.Disconnected);
    }

    private void secured_send(Protobuf.StatusMessage.Builder msg) {
        if (monitorActorState.expect(MonitorActorState.ConnectionState.Connected)) {
            send(msg);
        } else if (!monitorActorState.expect(MonitorActorState.ConnectionState.Disconnected)) {
            queue.tell(msg, getSender());
        }
    }

    private void send(Protobuf.StatusMessage.Builder msg) {
        msg.setUuid(ByteString.copyFrom(uuidToBytes(this.uuid)));
        msg.setMachineUuid(ByteString.copyFrom(uuidToBytes(this.machineUuid)));
        msg.setTimestamp(System.nanoTime());
        System.out.println(msg.build());
        connection.tell(msg.build(), getSelf());
    }

    private static InetSocketAddress parseAddress(String addrStr) throws MalformedURLException, UnknownHostException {
        // taken from http://stackoverflow.com/a/2346949
        String ipPattern = "(\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}):(\\d+)";
        String ipV6Pattern = "\\[([a-zA-Z0-9:]+)\\]:(\\d+)";
        Pattern p = Pattern.compile(ipPattern + "|" + ipV6Pattern);
        Matcher m = p.matcher(addrStr);
        if (m.matches()) {
            int port;
            InetAddress addr;
            if (m.group(1) != null) {
                // group(1) IP address, group(2) is port
                addr = Inet4Address.getByName(m.group(1));
                port = Integer.parseInt(m.group(2));
            } else if (m.group(3) != null) {
                // group(3) is IPv6 address, group(4) is port
                addr = Inet6Address.getByName(m.group(3));
                port = Integer.parseInt(m.group(4));
            } else {
                throw new MalformedURLException();
            }
            return new InetSocketAddress(addr, port);
        }
        throw new MalformedURLException();
    }

    // from http://stackoverflow.com/a/29836273
    private static UUID uuidFromBytes(byte[] bytes) {
        ByteBuffer bb = ByteBuffer.wrap(bytes);
        long firstLong = bb.getLong();
        long secondLong = bb.getLong();
        return new UUID(firstLong, secondLong);
    }

    // from http://stackoverflow.com/a/29836273
    private static byte[] uuidToBytes(UUID uuid) {
        ByteBuffer bb = ByteBuffer.wrap(new byte[16]);
        bb.putLong(uuid.getMostSignificantBits());
        bb.putLong(uuid.getLeastSignificantBits());
        return bb.array();
    }

    private void loadMachineUuid() {
        try {
            byte[] data = Files.readAllBytes(Paths.get(machineUuidFilePath));
            machineUuid = UUID.fromString(new String(data, "UTF-8"));
        } catch (IOException e) {
            // ignore error, file probably just does not exist
        }
    }

    private void saveMachineUuid() throws IOException {
        Files.write(Paths.get(machineUuidFilePath), machineUuid.toString().getBytes());
    }

    private static class StringListIterator implements Iterable<String> {

        private final String[] logs;

        public StringListIterator(String[] logs) {
            this.logs = logs;
        }

        // http://stackoverflow.com/a/14248833
        @Override
        public Iterator<String> iterator() {
            return new Iterator<String>() {
                private int pos = 0;

                public boolean hasNext() {
                    return logs.length > pos;
                }

                public String next() {
                    return logs[pos++];
                }

                public void remove() {
                    throw new UnsupportedOperationException("Cannot remove an element of an array.");
                }
            };
        }

        @Override
        public void forEach(Consumer<? super String> action) {
            throw new UnsupportedOperationException("Not implemented.");
        }

        @Override
        public Spliterator<String> spliterator() {
            throw new UnsupportedOperationException("Not implemented.");
        }
    }
}
