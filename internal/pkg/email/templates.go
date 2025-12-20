package email

const baseTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: #ffffff; }
        .header { background: #4F46E5; padding: 30px; text-align: center; }
        .header h1 { color: #ffffff; margin: 0; font-size: 24px; }
        .content { padding: 40px 30px; }
        .button { display: inline-block; padding: 14px 30px; background: #4F46E5; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; margin: 20px 0; }
        .button:hover { background: #4338CA; }
        .footer { padding: 30px; text-align: center; color: #666; font-size: 14px; background: #f9f9f9; }
        .note { background: #FEF3C7; padding: 15px; border-radius: 6px; margin: 20px 0; }
        .error { background: #FEE2E2; padding: 15px; border-radius: 6px; margin: 20px 0; color: #991B1B; }
        .warning { background: #FEF3C7; padding: 15px; border-radius: 6px; margin: 20px 0; color: #92400E; }
        .code { background: #F3F4F6; padding: 3px 8px; border-radius: 4px; font-family: monospace; }
    </style>
</head>
<body>
    <div class="container">
        {{.Content}}
    </div>
</body>
</html>
`

const welcomeTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: #ffffff; }
        .header { background: #4F46E5; padding: 30px; text-align: center; }
        .header h1 { color: #ffffff; margin: 0; font-size: 24px; }
        .content { padding: 40px 30px; }
        .button { display: inline-block; padding: 14px 30px; background: #4F46E5; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; margin: 20px 0; }
        .footer { padding: 30px; text-align: center; color: #666; font-size: 14px; background: #f9f9f9; }
        .feature { display: flex; margin: 15px 0; }
        .feature-icon { width: 40px; height: 40px; background: #EEF2FF; border-radius: 8px; display: flex; align-items: center; justify-content: center; margin-right: 15px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to {{.AppName}}! 🎉</h1>
        </div>
        <div class="content">
            <p>Hi {{.Name}},</p>
            <p>Thanks for signing up! We're excited to have you on board.</p>
            <p>{{.AppName}} helps you automate your workflows with ease. Here's what you can do:</p>
            <ul>
                <li><strong>Build workflows visually</strong> - No code required</li>
                <li><strong>Connect 100+ apps</strong> - Slack, Gmail, Notion, and more</li>
                <li><strong>Automate anything</strong> - Triggers, schedules, webhooks</li>
                <li><strong>AI-powered nodes</strong> - OpenAI, Claude, and more</li>
            </ul>
            <p style="text-align: center;">
                <a href="{{.LoginURL}}" class="button">Get Started</a>
            </p>
            <p>Need help? Check out our <a href="https://docs.linkflow.ai">documentation</a> or reach out to support.</p>
        </div>
        <div class="footer">
            <p>© {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const emailVerificationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: #ffffff; }
        .header { background: #4F46E5; padding: 30px; text-align: center; }
        .header h1 { color: #ffffff; margin: 0; font-size: 24px; }
        .content { padding: 40px 30px; }
        .button { display: inline-block; padding: 14px 30px; background: #4F46E5; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; margin: 20px 0; }
        .footer { padding: 30px; text-align: center; color: #666; font-size: 14px; background: #f9f9f9; }
        .note { background: #FEF3C7; padding: 15px; border-radius: 6px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Verify Your Email</h1>
        </div>
        <div class="content">
            <p>Hi {{.Name}},</p>
            <p>Please verify your email address by clicking the button below:</p>
            <p style="text-align: center;">
                <a href="{{.VerifyURL}}" class="button">Verify Email</a>
            </p>
            <div class="note">
                <strong>Note:</strong> This link will expire in {{.ExpiresIn}}.
            </div>
            <p>If you didn't create an account with {{.AppName}}, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>© {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const passwordResetTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: #ffffff; }
        .header { background: #4F46E5; padding: 30px; text-align: center; }
        .header h1 { color: #ffffff; margin: 0; font-size: 24px; }
        .content { padding: 40px 30px; }
        .button { display: inline-block; padding: 14px 30px; background: #4F46E5; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; margin: 20px 0; }
        .footer { padding: 30px; text-align: center; color: #666; font-size: 14px; background: #f9f9f9; }
        .note { background: #FEF3C7; padding: 15px; border-radius: 6px; margin: 20px 0; }
        .warning { background: #FEE2E2; padding: 15px; border-radius: 6px; margin: 20px 0; color: #991B1B; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Reset Your Password</h1>
        </div>
        <div class="content">
            <p>Hi {{.Name}},</p>
            <p>We received a request to reset your password. Click the button below to create a new password:</p>
            <p style="text-align: center;">
                <a href="{{.ResetURL}}" class="button">Reset Password</a>
            </p>
            <div class="note">
                <strong>Note:</strong> This link will expire in {{.ExpiresIn}}.
            </div>
            <div class="warning">
                <strong>Security tip:</strong> If you didn't request this password reset, please ignore this email and ensure your account is secure.
            </div>
        </div>
        <div class="footer">
            <p>© {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const workspaceInvitationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: #ffffff; }
        .header { background: #4F46E5; padding: 30px; text-align: center; }
        .header h1 { color: #ffffff; margin: 0; font-size: 24px; }
        .content { padding: 40px 30px; }
        .button { display: inline-block; padding: 14px 30px; background: #4F46E5; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; margin: 20px 0; }
        .footer { padding: 30px; text-align: center; color: #666; font-size: 14px; background: #f9f9f9; }
        .workspace-card { background: #F3F4F6; padding: 20px; border-radius: 8px; margin: 20px 0; text-align: center; }
        .workspace-name { font-size: 20px; font-weight: 600; color: #4F46E5; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>You're Invited!</h1>
        </div>
        <div class="content">
            <p>Hi there,</p>
            <p><strong>{{.InviterName}}</strong> has invited you to join their workspace on {{.AppName}}:</p>
            <div class="workspace-card">
                <div class="workspace-name">{{.WorkspaceName}}</div>
            </div>
            <p style="text-align: center;">
                <a href="{{.AcceptURL}}" class="button">Accept Invitation</a>
            </p>
            <p style="color: #666; font-size: 14px;">This invitation will expire in {{.ExpiresIn}}.</p>
        </div>
        <div class="footer">
            <p>© {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const executionFailedTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: #ffffff; }
        .header { background: #DC2626; padding: 30px; text-align: center; }
        .header h1 { color: #ffffff; margin: 0; font-size: 24px; }
        .content { padding: 40px 30px; }
        .button { display: inline-block; padding: 14px 30px; background: #4F46E5; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; margin: 20px 0; }
        .footer { padding: 30px; text-align: center; color: #666; font-size: 14px; background: #f9f9f9; }
        .error-box { background: #FEE2E2; padding: 20px; border-radius: 8px; margin: 20px 0; border-left: 4px solid #DC2626; }
        .error-message { font-family: monospace; font-size: 14px; color: #991B1B; white-space: pre-wrap; }
        .detail-row { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #E5E7EB; }
        .detail-label { color: #666; }
        .detail-value { font-weight: 500; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚠️ Workflow Execution Failed</h1>
        </div>
        <div class="content">
            <p>Your workflow <strong>{{.WorkflowName}}</strong> has failed.</p>
            
            <div class="detail-row">
                <span class="detail-label">Execution ID</span>
                <span class="detail-value">{{.ExecutionID}}</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Time</span>
                <span class="detail-value">{{.Timestamp}}</span>
            </div>
            
            <div class="error-box">
                <strong>Error Message:</strong>
                <div class="error-message">{{.ErrorMessage}}</div>
            </div>
            
            <p style="text-align: center;">
                <a href="{{.ViewURL}}" class="button">View Execution Details</a>
            </p>
        </div>
        <div class="footer">
            <p>© {{.AppName}}. All rights reserved.</p>
            <p style="font-size: 12px;">You're receiving this because you enabled execution failure notifications.</p>
        </div>
    </div>
</body>
</html>
`

const usageWarningTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: #ffffff; }
        .header { background: #F59E0B; padding: 30px; text-align: center; }
        .header h1 { color: #ffffff; margin: 0; font-size: 24px; }
        .content { padding: 40px 30px; }
        .button { display: inline-block; padding: 14px 30px; background: #4F46E5; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; margin: 20px 0; }
        .footer { padding: 30px; text-align: center; color: #666; font-size: 14px; background: #f9f9f9; }
        .progress-bar { background: #E5E7EB; border-radius: 10px; height: 20px; margin: 20px 0; overflow: hidden; }
        .progress-fill { background: #F59E0B; height: 100%; transition: width 0.3s; }
        .usage-text { text-align: center; font-size: 24px; font-weight: 600; color: #F59E0B; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚠️ Usage Warning</h1>
        </div>
        <div class="content">
            <p>Your workspace <strong>{{.WorkspaceName}}</strong> is approaching its {{.ResourceType}} limit.</p>
            
            <div class="usage-text">{{.UsagePercent}}% Used</div>
            <div class="progress-bar">
                <div class="progress-fill" style="width: {{.UsagePercent}}%"></div>
            </div>
            
            <p>To avoid service interruption, consider upgrading your plan or reducing usage.</p>
            
            <p style="text-align: center;">
                <a href="{{.UpgradeURL}}" class="button">Upgrade Plan</a>
            </p>
        </div>
        <div class="footer">
            <p>© {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const billingAlertTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: #ffffff; }
        .header { background: #4F46E5; padding: 30px; text-align: center; }
        .header h1 { color: #ffffff; margin: 0; font-size: 24px; }
        .content { padding: 40px 30px; }
        .button { display: inline-block; padding: 14px 30px; background: #4F46E5; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; margin: 20px 0; }
        .footer { padding: 30px; text-align: center; color: #666; font-size: 14px; background: #f9f9f9; }
        .alert-box { background: #FEF3C7; padding: 20px; border-radius: 8px; margin: 20px 0; border-left: 4px solid #F59E0B; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Billing Notification</h1>
        </div>
        <div class="content">
            {{if .PaymentFailed}}
            <div class="alert-box">
                <strong>Payment Failed</strong>
                <p>We were unable to process your payment. Please update your payment method to avoid service interruption.</p>
            </div>
            {{end}}
            
            {{if .SubscriptionEnding}}
            <p>Your subscription is ending on <strong>{{.EndDate}}</strong>.</p>
            <p>To continue using {{.AppName}}, please renew your subscription.</p>
            {{end}}
            
            {{if .InvoiceReady}}
            <p>Your invoice for <strong>{{.Amount}}</strong> is ready.</p>
            {{end}}
            
            <p style="text-align: center;">
                <a href="{{.BillingURL}}" class="button">Manage Billing</a>
            </p>
        </div>
        <div class="footer">
            <p>© {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`
