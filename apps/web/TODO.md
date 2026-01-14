# Project TODOs

## Security
- [ ] **Logout Improvement**: Implement token revocation on the backend.
    - Currently, logout is client-side only (removing the token from local storage).
    - **Goal**: When a user logs out, send a request to the backend to invalidate the Refresh Token (and potentially blacklist the Access Token until it expires).
    - **Backend**: Needs an endpoint `POST /api/v1/auth/logout`.
    - **Frontend**: Update `useAuthStore` or `ProfilePage` to call this endpoint before clearing the local token.

## Features
- [ ] **Matches Page**: Display a list of mutual likes.
- [ ] **Chat**: Real-time messaging with matches (WebSocket).
