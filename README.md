# Inga backend

## Email setup with Gmail

This project sends HTML emails through Gmail SMTP using a Google App Password.

### 1) Create a Google App Password

For a free Gmail account you usually need:
- 2-Step Verification enabled
- a Google App Password

### 2) Set email environment variables

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=yourname@gmail.com
SMTP_PASSWORD=your-google-app-password
SMTP_FROM_EMAIL=yourname@gmail.com
SMTP_FROM_NAME=Inga
SMTP_TEMPLATES_DIR=internal/notification/emails/templates
SMTP_USE_TLS=true
EMAIL_TO=recipient@example.com
```

### 3) Add HTML templates

Put your HTML files in:

```text
internal/notification/emails/templates
```

The starter template `welcome.html` is already included.

### 4) Run the email demo

```bash
go run ./cmd/email-demo
```

### 5) Send signup emails from the app

The auth signup flow uses the same mailer and sends `welcome.html` after a user is created.
