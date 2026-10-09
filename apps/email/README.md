# Email service

Run `docker compose -f devops/docker-compose.yaml up -d postgres nats`, apply migrations with `make migrate-up-user`, then run `make run-user` and `make run-email` from the repository root. The default `EMAIL_PROVIDER: fake` logs welcome emails without sending them.

For real delivery, set `EMAIL_PROVIDER=mailgun`, `EMAIL_FROM`, `MAILGUN_DOMAIN`, and `MAILGUN_API_KEY` in the email process environment. Set `MAILGUN_BASE_URL=https://api.eu.mailgun.net` for an EU region domain; the default is `https://api.mailgun.net`.

User creation commits the user and `user.created` event together. The user service retries publication to JetStream, and the email service subscribes to the event and acknowledges it only after the provider accepts it. Delivery is at least once: a crash after Mailgun accepts a request but before NATS records the acknowledgement can send the same welcome email again.

User requests and events use `first_name` and `last_name`. Each `user_outbox` message has its own `id`, with `user_id` referencing the user. Publication sets `processed_at` after JetStream acknowledges the message; processed rows are retained and excluded from subsequent batches.
