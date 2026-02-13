# P1 Friend Apply Status

```mermaid
stateDiagram-v2
    [*] --> Pending
    Pending --> Accepted: action=accept
    Pending --> Rejected: action=reject
    Pending --> Pending: duplicate submit blocked
    Accepted --> [*]
    Rejected --> [*]

    state Pending {
        [*] --> Unread
        Unread --> Read: MarkAsRead/async mark
        Read --> Read
    }
```
