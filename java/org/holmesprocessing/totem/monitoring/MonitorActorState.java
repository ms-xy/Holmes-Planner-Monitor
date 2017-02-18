package org.holmesprocessing.totem.monitoring;

public class MonitorActorState {

    public enum ConnectionState {
        Disconnected,
        Connecting,
        ExpectedHandshakeAck,
        Connected,
        Disconnecting,
    }
    private ConnectionState state;

    public MonitorActorState(ConnectionState initialState) {
        state = initialState;
    }

    public boolean set(ConnectionState next) {
        state = next;
        return true;
    }

    public boolean expect(ConnectionState prev) {
        return state == prev;
    }

    public boolean transition(ConnectionState prev, ConnectionState next) {
        return (expect(prev) || set(next));
    }

    @Override
    public String toString() {
        switch (state) {
            case Disconnected:
                return "Disconnected";
            case Connecting:
                return "Connecting";
            case ExpectedHandshakeAck:
                return "ExpectedHandshakeAck";
            case Connected:
                return "Connected";
            case Disconnecting:
                return "Disconnecting";
        }
        return "Undefined";
    }

}
