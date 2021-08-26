# Foreman worker for yggdrasil

This is a worker service for yggdrasil that can act as pull client for Foreman Remote Execution.

More info on [yggdrassil](https://github.com/RedHatInsights/yggdrasil).
The data message expected over mqtt to yggdrassil is standartized in [cloud-connector](https://github.com/RedHatInsights/cloud-connector#data)

## What does it do

It runs scripts.

We receive `content` from SmartProxy through yggdrasil, that contains a runnable script.
We write this script to a temp file, add executable permissions and run this script through `/bin/sh -c <script>`.

We scan outputs of this script and send them back through yggdrassil to defined `return_url`.
When the script finishes, we send exit_status to the same url.


## Worker specific API

We follow yggdrasil standards, so we expect developers to understand yggdrasil workflow.

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
