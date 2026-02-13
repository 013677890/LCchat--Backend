# P0 Device Business Status (User)

```mermaid
stateDiagram-v2
    [*] --> Offline
    Offline --> Online: UpdateDeviceStatus(0)
    Online --> Offline: UpdateDeviceStatus(1)
    Online --> Kicked: KickDevice
    Offline --> Kicked: KickDevice
    Online --> LoggedOut: Logout
    Offline --> LoggedOut: Logout
    Kicked --> Online: re-login
    LoggedOut --> Online: re-login
```
