# Discord webhook notifications

Contiwatch sends notification embeds with color `0x3498DB` (`3447003`). `SendEmbedWithLogo()` uses the message title as the webhook `username` and leaves the embed title empty. If `CONTIWATCH_PUBLIC_URL` is set, the avatar URL is `<PUBLIC_URL>/icons/contiwatch_logo_small.png`.

Sources: [Discord transport](../internal/notify/discord.go), [startup notification](../cmd/contiwatch/main.go), [scan/update notifications](../internal/server/server.go), and [configuration defaults](../internal/config/config.go).

## Global conditions

Runtime notifications require a non-empty `discord_webhook_url`, `discord_notifications_enabled=true`, and the relevant scenario switch. All Discord switches default to false, so unconfigured installations send no runtime notifications. The explicit test endpoint is independent of these saved switches.

See [configuration](configuration.md#secret-reads-and-updates) for hidden webhook reads, preservation, and clearing. Use an isolated webhook destination for delivery testing; it sends a real message.

## Webhook test

`POST /api/notifications/test` accepts a webhook URL in the request body rather than using the saved one. The endpoint requires the appropriate controller session or agent bearer credential and validates an official Discord HTTPS webhook URL.

```json
{"webhook_url": "https://discord.com/api/webhooks/<id>/<token>"}
```

The URL above is a placeholder; replace it only in a private request, never in committed examples. The resulting username is `Contiwatch test`, with description `Webhook verified.`.

## Controller startup

The controller sends `Contiwatch started` when the process starts and all global conditions plus `discord_notify_on_start=true` hold. Agents do not enter the controller startup-notification branch.

The description contains:

- Scheduler state: disabled, legacy interval (`Scheduler: enabled (every <duration>)`), Basic (`Scheduler: enabled (basic: <days> at <HH:MM>)`), or Cron (`Scheduler: enabled (cron: <expression>)`). Invalid plans are reported as invalid by the scheduler description helper.
- `Global policy: <policy>`.
- `Update stopped containers: <true|false>`.
- `Discord notifications: <true|false>`.
- `Remote servers: none configured` or a count and non-empty server names.
- `Local servers: none configured` or a count and non-empty server names.

See [scheduler configuration](configuration.md#scheduler) for plan formats and timezone behavior.

## Scan summary and automatic updates

After a server scan and its automatic updates, Contiwatch sends `Contiwatch updates` when either:

- `discord_notify_on_update_detected=true` and at least one update was detected; or
- `discord_notify_on_container_updated=true` and at least one container was updated.

The description includes `Server: <name> (local|remote)` and `Scanned images: <total>`. When detection notifications are enabled it adds `Updates detected: <detected>` and `Remaining outdated: <remaining>`. When update-result notifications are enabled it adds `Updated: <updated>`. Nonzero failures add `Failed: <failed>`.

The `Containers still outdated:` list is included when detection notifications are enabled and remaining outdated containers exist. The `Containers updated:` list is included when update-result notifications are enabled and updates succeeded.

Automatic local and remote updates are aggregated into this scan summary; they do not generate a separate notification for each container. Failures alone do not satisfy the summary's notification trigger when no update was detected or completed.

## Manual container update result

The manual update handler sends a per-container `Contiwatch updates` result when global conditions and `discord_notify_on_container_updated=true` hold and the operation reaches result processing. An error returned earlier by the handler is not a guaranteed per-container delivery.

```text
Server: <serverName>
Container: <containerName>
Result: <status>
Previous state: <previous>
Current state: <current>
```

`Result` is `updated` when `Updated=true`; otherwise it is the non-empty result `Message` (for example `update triggered; agent restarting`), or `not updated` if there is no message. Scheduled/automatic updates use the scan summary above.

## Payload and failures

```json
{
  "username": "Contiwatch started",
  "avatar_url": "https://contiwatch.example.com/icons/contiwatch_logo_small.png",
  "embeds": [{"description": "...", "color": 3447003}]
}
```

The avatar field is omitted when no public URL is configured. The transport uses a 10-second timeout; request/transport errors do not expose the webhook URL. Non-success HTTP responses are reported as notification failures. Delivery is not a durable queued/retried workflow: some runtime call sites ignore transport errors, while startup and explicit tests report failures. Use the explicit test to confirm intentional delivery, not as a production fixture.
