# P1 User Relation Status

```mermaid
stateDiagram-v2
    [*] --> None
    None --> Friend: accept apply
    Friend --> Deleted: delete friend
    Friend --> BlacklistFriend: add blacklist
    None --> BlacklistStranger: add blacklist
    BlacklistFriend --> Friend: remove blacklist
    BlacklistStranger --> Deleted: remove blacklist
    Deleted --> Friend: re-add friend
```
