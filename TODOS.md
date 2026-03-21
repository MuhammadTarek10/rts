# TODOs

## Forgot Password Backend Endpoint

**Priority:** P2
**Effort:** L
**Blocked by:** Email sending infrastructure (SMTP/SendGrid/etc.)

Implement `POST /api/v1/forgot-password` and `POST /api/v1/reset-password` endpoints in the auth service. Currently the frontend forgot-password page shows "coming soon" because no backend endpoint exists. Requires setting up email sending (e.g. Nodemailer + SMTP, or a service like SendGrid) to deliver reset links with time-limited tokens.
