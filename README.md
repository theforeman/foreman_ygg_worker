# Foreman worker for yggdrasil

This is a worker service for yggdrasil that can act as pull client for Foreman.

More info on [yggdrassil](https://github.com/RedHatInsights/yggdrasil).
The data message expected over mqtt to yggdrassil is standartized in [cloud-connector](https://github.com/RedHatInsights/cloud-connector#data)

## Dev docs

This worker expect to get message in following format:

```json
{
    "type": "data",
    "message_id": "a6a7d866-7de0-409a-84e0-3c56c4171bb7",
    "version": 1,
    "sent": "2021-01-12T15:30:08+00:00",
    "directive": "foreman",
    "metadata": {
        "return_url": "https://<foreman-proxy>/ssh/jobs/<job_id>/update"
    },
    "content": "https://<foreman-proxy>/ssh/jobs/<job_id>"
}

```

Where the `return_url` should expect to get `POST` requests in two formats:

```
{
  output: ['line_one', 'line_two'] # or single string with only line
  type: 'stdout' # or stderr
}
```

And last message will be

```
{
  exit_status: 0
}
```

## Hacking

We also have client in ruby for playing around in `ruby`
