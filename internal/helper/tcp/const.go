package tcp

type (
	TCP_TABLE_CLASS int32
	DWORD           uint32
	ULONG           uint32
	MIB_TCP_STATE   int32
)

const (
	TCP_TABLE_BASIC_LISTENER TCP_TABLE_CLASS = iota
	TCP_TABLE_BASIC_CONNECTIONS
	TCP_TABLE_BASIC_ALL
	TCP_TABLE_OWNER_PID_LISTENER
	TCP_TABLE_OWNER_PID_CONNECTIONS
	TCP_TABLE_OWNER_PID_ALL
	TCP_TABLE_OWNER_MODULE_LISTENER
	TCP_TABLE_OWNER_MODULE_CONNECTIONS
	TCP_TABLE_OWNER_MODULE_ALL
)

const (
	MIB_TCP_STATE_CLOSED     MIB_TCP_STATE = 1
	MIB_TCP_STATE_LISTEN     MIB_TCP_STATE = 2
	MIB_TCP_STATE_SYN_SENT   MIB_TCP_STATE = 3
	MIB_TCP_STATE_SYN_RCVD   MIB_TCP_STATE = 4
	MIB_TCP_STATE_ESTAB      MIB_TCP_STATE = 5
	MIB_TCP_STATE_FIN_WAIT1  MIB_TCP_STATE = 6
	MIB_TCP_STATE_FIN_WAIT2  MIB_TCP_STATE = 7
	MIB_TCP_STATE_CLOSE_WAIT MIB_TCP_STATE = 8
	MIB_TCP_STATE_CLOSING    MIB_TCP_STATE = 9
	MIB_TCP_STATE_LAST_ACK   MIB_TCP_STATE = 10
	MIB_TCP_STATE_TIME_WAIT  MIB_TCP_STATE = 11
	MIB_TCP_STATE_DELETE_TCB MIB_TCP_STATE = 12
)
