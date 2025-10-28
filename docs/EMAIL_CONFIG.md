## Environment Variables

Configure email functionality using either SendGrid (recommended) or SMTP:

### SendGrid (Recommended)
```env
SENDGRID_API_KEY=your-sendgrid-api-key
SMTP_FROM=noreply@webbuilder.com
BASE_URL=https://your-app.com
```

### SMTP (Alternative)
```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@example.com
SMTP_PASS=your-app-password
SMTP_FROM=noreply@webbuilder.com
BASE_URL=https://your-app.com
```

**Note:** If neither SendGrid nor SMTP credentials are configured, the system will run in mock mode and print email content to console instead of sending actual emails.
