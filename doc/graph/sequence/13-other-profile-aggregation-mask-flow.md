# P1 Other Profile Aggregation And Mask

```mermaid
sequenceDiagram
    participant C as Client
    participant G as Gateway.UserService
    participant U as User.UserService
    participant F as User.FriendService

    C->>G: GET /user/profile/:uuid
    par profile query
        G->>U: GetOtherProfile
    and relation query
        G->>F: CheckIsFriend
    end
    G->>G: if !isFriend then mask email/telephone
    G-->>C: merged profile response
```
