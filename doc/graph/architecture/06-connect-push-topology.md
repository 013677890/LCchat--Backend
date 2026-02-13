# P1 Connect Push Capability Topology

```mermaid
flowchart LR
    MsgSvc[Msg Service] -->|Push RPC| ConnectGRPC
    UserSvc[User Service] -->|Kick or status RPC| ConnectGRPC
    BizSvc[Other Biz Services] -->|Broadcast RPC| ConnectGRPC

    ConnectGRPC --> CM[ConnectionManager bucket index]
    CM --> D1[Device connection]
    CM --> D2[Device connection]
    CM --> DN[Device connection]
```
